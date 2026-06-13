package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type EventsHandler struct {
	logger *log.Logger
	pgPool *pgxpool.Pool
	redis  *redis.Client
}

type createEventRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	StartsAt    string   `json:"starts_at"`
	SeatRows    []string `json:"seat_rows"`
	SeatsPerRow int      `json:"seats_per_row"`
}

func NewEventsHandler(logger *log.Logger, pgPool *pgxpool.Pool, redisClient *redis.Client) *EventsHandler {
	return &EventsHandler{logger: logger, pgPool: pgPool, redis: redisClient}
}

func (h *EventsHandler) List(c echo.Context) error {
	ctx := c.Request().Context()
	cacheKey := "events:list:v1"

	if h.redis != nil {
		cached, err := h.redis.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			return c.Blob(http.StatusOK, "application/json", []byte(cached))
		}
	}

	rows, err := h.pgPool.Query(ctx, `SELECT id, title, starts_at FROM events ORDER BY starts_at ASC LIMIT 200`)
	if err != nil {
		h.logger.Printf("events list query failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "events_list_failed"})
	}
	defer rows.Close()

	items := make([]map[string]any, 0, 32)
	for rows.Next() {
		var id string
		var title string
		var startsAt time.Time
		if err := rows.Scan(&id, &title, &startsAt); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "events_list_failed"})
		}

		items = append(items, map[string]any{
			"id":        id,
			"title":     title,
			"starts_at": startsAt.UTC(),
		})
	}

	response := map[string]any{
		"items": items,
		"meta":  map[string]any{"cached": false},
	}

	if h.redis != nil {
		cachedResponse := map[string]any{
			"items": items,
			"meta":  map[string]any{"cached": true},
		}
		encoded, marshalErr := json.Marshal(cachedResponse)
		if marshalErr == nil {
			if err := h.redis.Set(ctx, cacheKey, encoded, 30*time.Second).Err(); err != nil {
				h.logger.Printf("events cache set failed: %v", err)
			}
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (h *EventsHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	_, userID, _, err := resolveAuthenticatedUser(ctx, c, h.pgPool)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing claims"})
	}

	var req createEventRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title_required"})
	}

	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid starts_at"})
	}

	seatRows := normalizeSeatRows(req.SeatRows)
	seatsPerRow := req.SeatsPerRow
	if seatsPerRow <= 0 || seatsPerRow > 100 {
		seatsPerRow = 20
	}

	tx, err := h.pgPool.Begin(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "transaction start failed"})
	}
	defer tx.Rollback(ctx)

	eventID := uuid.NewString()
	_, err = tx.Exec(
		ctx,
		`INSERT INTO events (id, title, description, starts_at)
		 VALUES ($1, $2, $3, $4)`,
		eventID,
		req.Title,
		strings.TrimSpace(req.Description),
		startsAt.UTC(),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "event_create_failed"})
	}

	seatsCreated := 0
	for _, seatRow := range seatRows {
		for number := 1; number <= seatsPerRow; number++ {
			_, err := tx.Exec(
				ctx,
				`INSERT INTO seats (id, event_id, seat_row, seat_number, status)
				 VALUES ($1, $2, $3, $4, 'available')`,
				uuid.NewString(),
				eventID,
				seatRow,
				number,
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "seat_create_failed"})
			}
			seatsCreated++
		}
	}

	_, _ = tx.Exec(
		ctx,
		`INSERT INTO audit_logs (id, actor_id, action, entity_type, entity_id, metadata)
		 VALUES ($1, $2, 'event.created', 'event', $3, jsonb_build_object('seats_created', $4))`,
		uuid.NewString(),
		userID,
		eventID,
		seatsCreated,
	)

	if err := tx.Commit(ctx); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "event_commit_failed"})
	}

	if h.redis != nil {
		if err := h.redis.Del(ctx, "events:list:v1").Err(); err != nil {
			h.logger.Printf("events cache invalidation failed: %v", err)
		}
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"id":            eventID,
		"message":       "created",
		"seats_created": seatsCreated,
	})
}

func (h *EventsHandler) Seats(c echo.Context) error {
	ctx := c.Request().Context()
	eventID := strings.TrimSpace(c.Param("eventId"))
	if eventID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "event_id_required"})
	}

	var exists int
	err := h.pgPool.QueryRow(ctx, `SELECT 1 FROM events WHERE id = $1`, eventID).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "event_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "event_lookup_failed"})
	}

	rows, err := h.pgPool.Query(
		ctx,
		`SELECT id, seat_row, seat_number, status
		 FROM seats
		 WHERE event_id = $1
		 ORDER BY seat_row ASC, seat_number ASC`,
		eventID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "seats_list_failed"})
	}
	defer rows.Close()

	seats := make([]map[string]any, 0, 128)
	for rows.Next() {
		var seatID string
		var seatRow string
		var seatNumber int
		var status string
		if err := rows.Scan(&seatID, &seatRow, &seatNumber, &status); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "seats_list_failed"})
		}

		seats = append(seats, map[string]any{
			"id":     seatID,
			"row":    seatRow,
			"number": seatNumber,
			"status": status,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"event_id": eventID,
		"seats":    seats,
	})
}

func normalizeSeatRows(input []string) []string {
	if len(input) == 0 {
		return []string{"A", "B", "C", "D", "E"}
	}

	rows := make([]string, 0, len(input))
	seen := map[string]struct{}{}
	for _, raw := range input {
		value := strings.ToUpper(strings.TrimSpace(raw))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		rows = append(rows, value)
	}

	if len(rows) == 0 {
		return []string{"A", "B", "C", "D", "E"}
	}

	return rows
}
