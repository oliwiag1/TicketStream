package services

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
)

const (
	outboxExchangeName = "ticketstream.events"
	outboxBatchSize    = 25
)

type OutboxService struct {
	pgPool *pgxpool.Pool
	ch     *amqp091.Channel
	logger *log.Logger
}

type outboxRecord struct {
	ID        string
	EventType string
	Payload   []byte
}

func NewOutboxService(pgPool *pgxpool.Pool, ch *amqp091.Channel, logger *log.Logger) *OutboxService {
	return &OutboxService{pgPool: pgPool, ch: ch, logger: logger}
}

func (s *OutboxService) PublishPending(ctx context.Context) error {
	if s.pgPool == nil || s.ch == nil {
		return nil
	}

	if err := s.ch.ExchangeDeclare(outboxExchangeName, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare outbox exchange: %w", err)
	}

	tx, err := s.pgPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin outbox tx: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(
		ctx,
		`SELECT id, event_type, payload_json::text
		 FROM outbox_events
		 WHERE published_at IS NULL
		 ORDER BY created_at
		 FOR UPDATE SKIP LOCKED
		 LIMIT $1`,
		outboxBatchSize,
	)
	if err != nil {
		return fmt.Errorf("query outbox rows: %w", err)
	}
	defer rows.Close()

	records := make([]outboxRecord, 0, outboxBatchSize)
	for rows.Next() {
		var record outboxRecord
		var payload string
		if err := rows.Scan(&record.ID, &record.EventType, &payload); err != nil {
			return fmt.Errorf("scan outbox row: %w", err)
		}
		record.Payload = []byte(payload)
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate outbox rows: %w", err)
	}

	for _, record := range records {
		routingKey, err := routingKeyForEvent(record.EventType)
		if err != nil {
			return err
		}

		if err := s.ch.PublishWithContext(
			ctx,
			outboxExchangeName,
			routingKey,
			false,
			false,
			amqp091.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp091.Persistent,
				Body:         record.Payload,
			},
		); err != nil {
			return fmt.Errorf("publish outbox event %s: %w", record.ID, err)
		}

		if _, err := tx.Exec(ctx, `UPDATE outbox_events SET published_at = NOW() WHERE id = $1`, record.ID); err != nil {
			return fmt.Errorf("mark outbox event published %s: %w", record.ID, err)
		}
	}

	if len(records) > 0 {
		s.logger.Printf("published %d outbox event(s)", len(records))
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit outbox tx: %w", err)
	}
	return nil
}

func routingKeyForEvent(eventType string) (string, error) {
	switch eventType {
	case "payment.succeeded":
		return "ticket.generated", nil
	default:
		return "", fmt.Errorf("unsupported outbox event type: %s", eventType)
	}
}
