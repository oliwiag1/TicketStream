package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"ticketstream/backend/internal/config"
	"ticketstream/backend/internal/observability"
	"ticketstream/backend/internal/services"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type ReservationHandler struct {
	cfg      config.Config
	logger   *log.Logger
	pgPool   *pgxpool.Pool
	redis    *redis.Client
	realtime *services.RealtimeService
}

type reserveRequest struct {
	SeatID string `json:"seat_id"`
}

func NewReservationHandler(cfg config.Config, logger *log.Logger, pgPool *pgxpool.Pool, redisClient *redis.Client, realtime *services.RealtimeService) *ReservationHandler {
	return &ReservationHandler{cfg: cfg, logger: logger, pgPool: pgPool, redis: redisClient, realtime: realtime}
}

func (h *ReservationHandler) Reserve(c echo.Context) error {
	ctx := c.Request().Context()
	subject, userID, _, err := resolveAuthenticatedUser(ctx, c, h.pgPool)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing claims"})
	}

	eventID := c.Param("eventId")
	var req reserveRequest
	if err := c.Bind(&req); err != nil || req.SeatID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	lockTTL := time.Duration(h.cfg.SeatLockTTLSeconds) * time.Second
	lockKey := seatLockKey(eventID, req.SeatID)
	lockOk, err := h.redis.SetNX(ctx, lockKey, subject, lockTTL).Result()
	if err != nil {
		h.logger.Printf("reserve redis lock error event=%s seat=%s err=%v", eventID, req.SeatID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "lock unavailable"})
	}
	if !lockOk {
		observability.IncLockContention()
		return c.JSON(http.StatusConflict, map[string]string{"error": "seat_already_reserved"})
	}

	tx, err := h.pgPool.Begin(ctx)
	if err != nil {
		h.releaseSeatLock(ctx, lockKey, subject)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "transaction start failed"})
	}
	defer tx.Rollback(ctx)

	var lockedSeatID string
	err = tx.QueryRow(
		ctx,
		`UPDATE seats
		 SET status = 'locked', updated_at = NOW()
		 WHERE id = $1 AND event_id = $2 AND status = 'available'
		 RETURNING id`,
		req.SeatID,
		eventID,
	).Scan(&lockedSeatID)
	if err != nil {
		h.releaseSeatLock(ctx, lockKey, subject)
		if err == pgx.ErrNoRows {
			observability.IncLockContention()
			return c.JSON(http.StatusConflict, map[string]string{"error": "seat_not_available"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "seat lock failed"})
	}

	reservationID := uuid.NewString()
	expiresAt := time.Now().UTC().Add(lockTTL)

	_, err = tx.Exec(
		ctx,
		`INSERT INTO reservations (id, user_id, event_id, seat_id, status, expires_at)
		 VALUES ($1, $2, $3, $4, 'locked', $5)`,
		reservationID,
		userID,
		eventID,
		req.SeatID,
		expiresAt,
	)
	if err != nil {
		h.releaseSeatLock(ctx, lockKey, subject)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "reservation create failed"})
	}

	_, _ = tx.Exec(
		ctx,
		`INSERT INTO audit_logs (id, actor_id, action, entity_type, entity_id, metadata)
		 VALUES ($1, $2, 'reserve.locked', 'reservation', $3, jsonb_build_object('event_id', $4::text, 'seat_id', $5::text))`,
		uuid.NewString(),
		userID,
		reservationID,
		eventID,
		req.SeatID,
	)

	if err := tx.Commit(ctx); err != nil {
		h.releaseSeatLock(ctx, lockKey, subject)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "reservation commit failed"})
	}

	if h.realtime != nil {
		if err := h.realtime.BroadcastSeatUpdate(ctx, eventID, lockedSeatID, "locked", time.Now().UTC()); err != nil {
			h.logger.Printf("realtime reserve broadcast failed event=%s seat=%s err=%v", eventID, lockedSeatID, err)
		}
	}

	return c.JSON(http.StatusAccepted, map[string]any{
		"reservation_id": reservationID,
		"event_id":       eventID,
		"seat_id":        lockedSeatID,
		"status":         "locked",
		"expires_at":     expiresAt,
	})
}

func (h *ReservationHandler) Cancel(c echo.Context) error {
	ctx := c.Request().Context()
	subject, userID, _, err := resolveAuthenticatedUser(ctx, c, h.pgPool)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing claims"})
	}

	reservationID := c.Param("reservationId")
	tx, err := h.pgPool.Begin(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "transaction start failed"})
	}
	defer tx.Rollback(ctx)

	var seatID, eventID, status string
	err = tx.QueryRow(
		ctx,
		`SELECT seat_id, event_id, status
		 FROM reservations
		 WHERE id = $1 AND user_id = $2
		 FOR UPDATE`,
		reservationID,
		userID,
	).Scan(&seatID, &eventID, &status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "reservation_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "reservation fetch failed"})
	}

	if status == "sold" {
		return c.JSON(http.StatusConflict, map[string]string{"error": "reservation_already_paid"})
	}

	_, err = tx.Exec(ctx, `UPDATE reservations SET status = 'released' WHERE id = $1`, reservationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "reservation release failed"})
	}

	_, err = tx.Exec(ctx, `UPDATE seats SET status = 'available', updated_at = NOW() WHERE id = $1 AND status = 'locked'`, seatID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "seat release failed"})
	}

	_, _ = tx.Exec(
		ctx,
		`INSERT INTO audit_logs (id, actor_id, action, entity_type, entity_id, metadata)
		 VALUES ($1, $2, 'reserve.released', 'reservation', $3, jsonb_build_object('event_id', $4::text, 'seat_id', $5::text))`,
		uuid.NewString(),
		userID,
		reservationID,
		eventID,
		seatID,
	)

	if err := tx.Commit(ctx); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "reservation commit failed"})
	}

	h.releaseSeatLock(ctx, seatLockKey(eventID, seatID), subject)

	if h.realtime != nil {
		if err := h.realtime.BroadcastSeatUpdate(ctx, eventID, seatID, "available", time.Now().UTC()); err != nil {
			h.logger.Printf("realtime release broadcast failed event=%s seat=%s err=%v", eventID, seatID, err)
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message":        "released",
		"reservation_id": c.Param("reservationId"),
	})
}

func seatLockKey(eventID, seatID string) string {
	return "seat_lock:" + eventID + ":" + seatID
}

func (h *ReservationHandler) releaseSeatLock(ctx context.Context, key, subject string) {
	script := `if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("DEL", KEYS[1]) else return 0 end`
	_ = h.redis.Eval(ctx, script, []string{key}, subject).Err()
}
