package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"ticketstream/backend/internal/services"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type PaymentHandler struct {
	logger   *log.Logger
	pgPool   *pgxpool.Pool
	redis    *redis.Client
	realtime *services.RealtimeService
}

type payRequest struct {
	ReservationID string `json:"reservation_id"`
	PaymentMethod string `json:"payment_method"`
}

func NewPaymentHandler(logger *log.Logger, pgPool *pgxpool.Pool, redisClient *redis.Client, realtime *services.RealtimeService) *PaymentHandler {
	return &PaymentHandler{logger: logger, pgPool: pgPool, redis: redisClient, realtime: realtime}
}

func (h *PaymentHandler) Pay(c echo.Context) error {
	ctx := c.Request().Context()
	subject, userID, _, err := resolveAuthenticatedUser(ctx, c, h.pgPool)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing claims"})
	}

	idempotencyKey := c.Request().Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing idempotency key"})
	}

	var req payRequest
	if err := c.Bind(&req); err != nil || req.ReservationID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	storedCode, storedBody, found, err := h.lookupIdempotency(ctx, idempotencyKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "idempotency lookup failed"})
	}
	if found {
		return c.Blob(storedCode, "application/json", storedBody)
	}

	tx, err := h.pgPool.Begin(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "transaction start failed"})
	}
	defer tx.Rollback(ctx)

	var eventID, seatID, reservationStatus, seatStatus string
	var expiresAt *time.Time
	err = tx.QueryRow(
		ctx,
		`SELECT r.event_id, r.seat_id, r.status, r.expires_at, s.status
		 FROM reservations r
		 JOIN seats s ON s.id = r.seat_id
		 WHERE r.id = $1 AND r.user_id = $2
		 FOR UPDATE OF r, s`,
		req.ReservationID,
		userID,
	).Scan(&eventID, &seatID, &reservationStatus, &expiresAt, &seatStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "reservation_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "reservation fetch failed"})
	}

	if reservationStatus == "sold" {
		body, _ := json.Marshal(map[string]any{"status": "accepted", "reservation_id": req.ReservationID, "already_paid": true})
		if err := h.persistIdempotency(ctx, tx, idempotencyKey, http.StatusOK, body); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "idempotency persist failed"})
		}
		if err := tx.Commit(ctx); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "payment commit failed"})
		}
		return c.Blob(http.StatusOK, "application/json", body)
	}

	if reservationStatus != "locked" || seatStatus != "locked" {
		return c.JSON(http.StatusConflict, map[string]string{"error": "reservation_conflict"})
	}

	if expiresAt != nil && time.Now().UTC().After(expiresAt.UTC()) {
		return c.JSON(http.StatusConflict, map[string]string{"error": "reservation_expired"})
	}

	commandTag, err := tx.Exec(
		ctx,
		`UPDATE seats
		 SET status = 'sold', updated_at = NOW()
		 WHERE id = $1 AND status = 'locked'`,
		seatID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "seat update failed"})
	}
	if commandTag.RowsAffected() != 1 {
		return c.JSON(http.StatusConflict, map[string]string{"error": "seat_not_locked"})
	}

	_, err = tx.Exec(ctx, `UPDATE reservations SET status = 'sold' WHERE id = $1`, req.ReservationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "reservation update failed"})
	}

	outboxPayload, _ := json.Marshal(map[string]any{
		"reservation_id": req.ReservationID,
		"event_id":       eventID,
		"seat_id":        seatID,
		"user_id":        userID,
		"payment_method": req.PaymentMethod,
	})

	_, err = tx.Exec(
		ctx,
		`INSERT INTO outbox_events (id, aggregate_id, event_type, payload_json)
		 VALUES ($1, $2, 'payment.succeeded', $3::jsonb)`,
		uuid.NewString(),
		req.ReservationID,
		string(outboxPayload),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "outbox insert failed"})
	}

	_, _ = tx.Exec(
		ctx,
		`INSERT INTO audit_logs (id, actor_id, action, entity_type, entity_id, metadata)
		 VALUES ($1, $2, 'payment.accepted', 'reservation', $3, jsonb_build_object('event_id', $4::text, 'seat_id', $5::text))`,
		uuid.NewString(),
		userID,
		req.ReservationID,
		eventID,
		seatID,
	)

	responseBody, _ := json.Marshal(map[string]any{
		"status":         "accepted",
		"reservation_id": req.ReservationID,
	})

	if err := h.persistIdempotency(ctx, tx, idempotencyKey, http.StatusOK, responseBody); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "idempotency persist failed"})
	}

	if err := tx.Commit(ctx); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "payment commit failed"})
	}

	h.releaseSeatLock(ctx, seatLockKey(eventID, seatID), subject)

	if h.realtime != nil {
		if err := h.realtime.BroadcastSeatUpdate(ctx, eventID, seatID, "sold", time.Now().UTC()); err != nil {
			h.logger.Printf("realtime payment broadcast failed event=%s seat=%s err=%v", eventID, seatID, err)
		}
	}

	return c.Blob(http.StatusOK, "application/json", responseBody)
}

func (h *PaymentHandler) lookupIdempotency(ctx context.Context, key string) (int, []byte, bool, error) {
	var code int
	var bodyText string
	err := h.pgPool.QueryRow(
		ctx,
		`SELECT response_code, response_body::text FROM idempotency_requests WHERE idempotency_key = $1`,
		key,
	).Scan(&code, &bodyText)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil, false, nil
		}
		return 0, nil, false, err
	}
	return code, []byte(bodyText), true, nil
}

func (h *PaymentHandler) persistIdempotency(ctx context.Context, tx pgx.Tx, key string, responseCode int, responseBody []byte) error {
	_, err := tx.Exec(
		ctx,
		`INSERT INTO idempotency_requests (id, idempotency_key, response_code, response_body)
		 VALUES ($1, $2, $3, $4::jsonb)
		 ON CONFLICT (idempotency_key) DO NOTHING`,
		uuid.NewString(),
		key,
		responseCode,
		string(responseBody),
	)
	return err
}

func (h *PaymentHandler) releaseSeatLock(ctx context.Context, key, subject string) {
	if h.redis == nil {
		return
	}
	script := `if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("DEL", KEYS[1]) else return 0 end`
	_ = h.redis.Eval(ctx, script, []string{key}, subject).Err()
}
