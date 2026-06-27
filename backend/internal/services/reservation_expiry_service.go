package services

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ReservationExpiryService struct {
	logger   *log.Logger
	pgPool   *pgxpool.Pool
	redis    *redis.Client
	realtime *RealtimeService
}

type expiredReservation struct {
	reservationID string
	eventID       string
	seatID        string
}

func NewReservationExpiryService(logger *log.Logger, pgPool *pgxpool.Pool, redisClient *redis.Client, realtime *RealtimeService) *ReservationExpiryService {
	return &ReservationExpiryService{logger: logger, pgPool: pgPool, redis: redisClient, realtime: realtime}
}

func (s *ReservationExpiryService) ReleaseExpired(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = 100
	}

	tx, err := s.pgPool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(
		ctx,
		`SELECT id, event_id, seat_id
		 FROM reservations
		 WHERE status = 'locked'
		   AND expires_at IS NOT NULL
		   AND expires_at <= NOW()
		 ORDER BY expires_at
		 FOR UPDATE SKIP LOCKED
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	items := make([]expiredReservation, 0, limit)
	for rows.Next() {
		var item expiredReservation
		if err := rows.Scan(&item.reservationID, &item.eventID, &item.seatID); err != nil {
			return 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	for _, item := range items {
		if _, err := tx.Exec(ctx, `UPDATE reservations SET status = 'released' WHERE id = $1 AND status = 'locked'`, item.reservationID); err != nil {
			return 0, err
		}
		if _, err := tx.Exec(ctx, `UPDATE seats SET status = 'available', updated_at = NOW() WHERE id = $1 AND status = 'locked'`, item.seatID); err != nil {
			return 0, err
		}

		_, _ = tx.Exec(
			ctx,
			`INSERT INTO audit_logs (id, action, entity_type, entity_id, metadata)
			 VALUES ($1, 'reserve.expired_release', 'reservation', $2, jsonb_build_object('event_id', $3::text, 'seat_id', $4::text))`,
			uuid.NewString(),
			item.reservationID,
			item.eventID,
			item.seatID,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	for _, item := range items {
		if s.redis != nil {
			_ = s.redis.Del(ctx, seatLockKey(item.eventID, item.seatID)).Err()
		}
		if s.realtime != nil {
			if err := s.realtime.BroadcastSeatUpdate(ctx, item.eventID, item.seatID, "available", time.Now().UTC()); err != nil {
				s.logger.Printf("realtime expired release broadcast failed event=%s seat=%s err=%v", item.eventID, item.seatID, err)
			}
		}
	}

	if len(items) > 0 {
		s.logger.Printf("released %d expired reservation(s)", len(items))
	}

	return len(items), nil
}

func seatLockKey(eventID, seatID string) string {
	return "seat_lock:" + eventID + ":" + seatID
}
