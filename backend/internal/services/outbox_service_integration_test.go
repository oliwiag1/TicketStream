//go:build integration

package services

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ticketstream/backend/internal/worker/queue"
	"ticketstream/backend/pkg/broker"
	"ticketstream/backend/pkg/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
)

func TestIntegrationPublishPendingPublishesAndMarksOutbox(t *testing.T) {
	ctx := context.Background()
	pgPool := integrationPostgres(t)
	defer pgPool.Close()

	conn, ch := integrationRabbitMQ(t)
	defer conn.Close()
	defer ch.Close()

	purgeQueue(t, ch, queue.TicketQueueName)
	purgeQueue(t, ch, queue.TicketDLQName)

	reservationID := uuid.NewString()
	payload := `{"reservation_id":"` + reservationID + `","event_id":"event-1","seat_id":"seat-1","user_id":"user-1","payment_method":"card"}`
	_, err := pgPool.Exec(
		ctx,
		`INSERT INTO outbox_events (id, aggregate_id, event_type, payload_json)
		 VALUES ($1, $2, 'payment.succeeded', $3::jsonb)`,
		uuid.NewString(),
		reservationID,
		payload,
	)
	if err != nil {
		t.Fatalf("insert outbox event: %v", err)
	}

	service := NewOutboxService(pgPool, ch, log.New(io.Discard, "", 0))
	if err := service.PublishPending(ctx); err != nil {
		t.Fatalf("PublishPending() error = %v", err)
	}

	msg, ok, err := ch.Get(queue.TicketQueueName, true)
	if err != nil {
		t.Fatalf("get published message: %v", err)
	}
	if !ok {
		t.Fatalf("expected published message in queue %s", queue.TicketQueueName)
	}
	if string(msg.Body) != payload {
		t.Fatalf("published payload = %s, want %s", string(msg.Body), payload)
	}

	var publishedAt *time.Time
	if err := pgPool.QueryRow(
		ctx,
		`SELECT published_at FROM outbox_events WHERE aggregate_id = $1`,
		reservationID,
	).Scan(&publishedAt); err != nil {
		t.Fatalf("query published_at: %v", err)
	}
	if publishedAt == nil {
		t.Fatalf("published_at = nil, want timestamp")
	}
}

func integrationPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	postgresURL := os.Getenv("TEST_POSTGRES_URL")
	if postgresURL == "" {
		postgresURL = "postgres://ticketstream:ticketstream@localhost:5433/ticketstream?sslmode=disable"
	}

	pgPool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		t.Skipf("skipping integration test: cannot create postgres pool: %v", err)
	}
	if err := pgPool.Ping(ctx); err != nil {
		pgPool.Close()
		t.Skipf("skipping integration test: postgres unavailable: %v", err)
	}

	migrationsDir := os.Getenv("TEST_MIGRATIONS_DIR")
	if migrationsDir == "" {
		wd, wdErr := os.Getwd()
		if wdErr != nil {
			pgPool.Close()
			t.Skipf("skipping integration test: cannot resolve working dir: %v", wdErr)
		}
		migrationsDir = filepath.Clean(filepath.Join(wd, "..", "..", "..", "migrations"))
	}

	if err := db.RunMigrations(ctx, pgPool, migrationsDir, log.New(io.Discard, "", 0)); err != nil {
		pgPool.Close()
		t.Skipf("skipping integration test: migrations failed: %v", err)
	}

	return pgPool
}

func integrationRabbitMQ(t *testing.T) (*amqp091.Connection, *amqp091.Channel) {
	t.Helper()

	rabbitURL := os.Getenv("TEST_RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, ch, err := broker.NewRabbitMQ(rabbitURL)
	if err != nil {
		t.Skipf("skipping integration test: rabbitmq unavailable: %v", err)
	}

	if err := ch.ExchangeDeclare(queue.ExchangeName, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		t.Skipf("skipping integration test: exchange declare failed: %v", err)
	}
	if _, err := ch.QueueDeclare(queue.TicketQueueName, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		t.Skipf("skipping integration test: queue declare failed: %v", err)
	}
	if _, err := ch.QueueDeclare(queue.TicketDLQName, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		t.Skipf("skipping integration test: dlq declare failed: %v", err)
	}
	if err := ch.QueueBind(queue.TicketQueueName, queue.TicketRoutingKey, queue.ExchangeName, false, nil); err != nil {
		ch.Close()
		conn.Close()
		t.Skipf("skipping integration test: queue bind failed: %v", err)
	}
	if err := ch.QueueBind(queue.TicketDLQName, queue.TicketDLQRouteKey, queue.ExchangeName, false, nil); err != nil {
		ch.Close()
		conn.Close()
		t.Skipf("skipping integration test: dlq bind failed: %v", err)
	}

	return conn, ch
}

func purgeQueue(t *testing.T, ch *amqp091.Channel, queueName string) {
	t.Helper()
	if _, err := ch.QueuePurge(queueName, false); err != nil {
		t.Fatalf("purge queue %s: %v", queueName, err)
	}
}
