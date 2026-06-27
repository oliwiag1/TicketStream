package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type TicketsHandler struct {
	logger *log.Logger
	pgPool *pgxpool.Pool
}

func NewTicketsHandler(logger *log.Logger, pgPool *pgxpool.Pool) *TicketsHandler {
	return &TicketsHandler{logger: logger, pgPool: pgPool}
}

func (h *TicketsHandler) ListMine(c echo.Context) error {
	ctx := c.Request().Context()
	_, userID, _, err := resolveAuthenticatedUser(ctx, c, h.pgPool)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing claims"})
	}

	rows, err := h.pgPool.Query(
		ctx,
		`SELECT
			r.id,
			e.id,
			e.title,
			e.starts_at,
			s.id,
			s.seat_row,
			s.seat_number,
			r.status,
			COALESCE(
				MAX(CASE WHEN a.action = 'payment.accepted' THEN a.created_at END),
				r.created_at
			) AS purchased_at
		 FROM reservations r
		 JOIN events e ON e.id = r.event_id
		 JOIN seats s ON s.id = r.seat_id
		 LEFT JOIN audit_logs a
			ON a.entity_type = 'reservation'
			AND a.entity_id = r.id
			AND a.action = 'payment.accepted'
		 WHERE r.user_id = $1
			AND r.status = 'sold'
		 GROUP BY r.id, e.id, e.title, e.starts_at, s.id, s.seat_row, s.seat_number, r.status, r.created_at
		 ORDER BY purchased_at DESC
		 LIMIT 200`,
		userID,
	)
	if err != nil {
		h.logger.Printf("tickets list query failed user_id=%s err=%v", userID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "tickets_list_failed"})
	}
	defer rows.Close()

	items := make([]map[string]any, 0, 16)
	for rows.Next() {
		var (
			reservationID string
			eventID       string
			eventTitle    string
			eventStartsAt time.Time
			seatID        string
			seatRow       string
			seatNumber    int
			status        string
			purchasedAt   time.Time
		)

		if err := rows.Scan(
			&reservationID,
			&eventID,
			&eventTitle,
			&eventStartsAt,
			&seatID,
			&seatRow,
			&seatNumber,
			&status,
			&purchasedAt,
		); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "tickets_list_failed"})
		}

		items = append(items, map[string]any{
			"reservation_id":  reservationID,
			"event_id":        eventID,
			"event_title":     eventTitle,
			"event_starts_at": eventStartsAt.UTC(),
			"seat_id":         seatID,
			"seat_row":        seatRow,
			"seat_number":     seatNumber,
			"status":          status,
			"purchased_at":    purchasedAt.UTC(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"items": items,
	})
}
