//go:build integration

package queue

import (
	"context"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"ticketstream/backend/pkg/broker"

	"github.com/rabbitmq/amqp091-go"
)

func TestIntegrationConsumerRoutesMessagesToMainQueueAndDLQ(t *testing.T) {
	conn, workerCh := integrationRabbitMQ(t)
	defer conn.Close()
	defer workerCh.Close()

	testCh, err := conn.Channel()
	if err != nil {
		t.Fatalf("open rabbitmq test channel: %v", err)
	}
	defer testCh.Close()

	consumer := NewConsumer(workerCh, log.New(io.Discard, "", 0))
	go func() {
		_ = consumer.Start()
	}()

	waitForQueueReady(t, testCh, TicketQueueName)
	waitForQueueReady(t, testCh, TicketDLQName)
	purgeQueue(t, testCh, TicketQueueName)
	purgeQueue(t, testCh, TicketDLQName)

	validPayload := []byte(`{"reservation_id":"r1","event_id":"e1","seat_id":"s1","user_id":"u1","payment_method":"card"}`)
	if err := testCh.PublishWithContext(
		context.Background(),
		ExchangeName,
		TicketRoutingKey,
		false,
		false,
		amqp091.Publishing{ContentType: "application/json", DeliveryMode: amqp091.Persistent, Body: validPayload},
	); err != nil {
		t.Fatalf("publish valid payload: %v", err)
	}

	waitForDrain(t, testCh, TicketQueueName)
	if dlqCount := messageCount(t, testCh, TicketDLQName); dlqCount != 0 {
		t.Fatalf("dlq count after valid payload = %d, want 0", dlqCount)
	}

	invalidPayload := []byte(`{"event_id":"e1","seat_id":"s1","user_id":"u1"}`)
	if err := testCh.PublishWithContext(
		context.Background(),
		ExchangeName,
		TicketRoutingKey,
		false,
		false,
		amqp091.Publishing{ContentType: "application/json", DeliveryMode: amqp091.Persistent, Body: invalidPayload},
	); err != nil {
		t.Fatalf("publish invalid payload: %v", err)
	}

	waitForMessage(t, testCh, TicketDLQName)
	dlqMsg, ok, err := testCh.Get(TicketDLQName, true)
	if err != nil {
		t.Fatalf("get dlq message: %v", err)
	}
	if !ok {
		t.Fatalf("expected invalid payload in DLQ")
	}
	if string(dlqMsg.Body) != string(invalidPayload) {
		t.Fatalf("dlq body = %s, want %s", string(dlqMsg.Body), string(invalidPayload))
	}
}

func TestIntegrationConsumerRetriesTransientMessagesBeforeDLQ(t *testing.T) {
	conn, workerCh := integrationRabbitMQ(t)
	defer conn.Close()
	defer workerCh.Close()

	testCh, err := conn.Channel()
	if err != nil {
		t.Fatalf("open rabbitmq test channel: %v", err)
	}
	defer testCh.Close()

	consumer := NewConsumer(workerCh, log.New(io.Discard, "", 0))
	go func() {
		_ = consumer.Start()
	}()

	waitForQueueReady(t, testCh, TicketQueueName)
	waitForQueueReady(t, testCh, TicketDLQName)
	purgeQueue(t, testCh, TicketQueueName)
	purgeQueue(t, testCh, TicketDLQName)

	payload := []byte(`{"reservation_id":"r1","event_id":"e1","seat_id":"s1","user_id":"u1","payment_method":"simulate_transient"}`)
	if err := testCh.PublishWithContext(
		context.Background(),
		ExchangeName,
		TicketRoutingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         payload,
		},
	); err != nil {
		t.Fatalf("publish retried payload: %v", err)
	}

	waitForMessage(t, testCh, TicketDLQName)
	msg, ok, err := testCh.Get(TicketDLQName, true)
	if err != nil {
		t.Fatalf("get retried dlq message: %v", err)
	}
	if !ok {
		t.Fatalf("expected retried payload in DLQ")
	}
	retryCount := retryCountFromHeaders(msg.Headers)
	if retryCount != MaxRetryAttempts {
		t.Fatalf("dlq retry_count = %d, want %d", retryCount, MaxRetryAttempts)
	}
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

	return conn, ch
}

func waitForQueueReady(t *testing.T, ch *amqp091.Channel, queueName string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		queue, err := ch.QueueInspect(queueName)
		if err == nil && queue.Name == queueName {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("queue %s not ready", queueName)
}

func waitForDrain(t *testing.T, ch *amqp091.Channel, queueName string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if messageCount(t, ch, queueName) == 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("queue %s did not drain in time", queueName)
}

func waitForMessage(t *testing.T, ch *amqp091.Channel, queueName string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if messageCount(t, ch, queueName) > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("queue %s did not receive message in time", queueName)
}

func messageCount(t *testing.T, ch *amqp091.Channel, queueName string) int {
	t.Helper()
	queue, err := ch.QueueInspect(queueName)
	if err != nil {
		t.Fatalf("inspect queue %s: %v", queueName, err)
	}
	return queue.Messages
}

func purgeQueue(t *testing.T, ch *amqp091.Channel, queueName string) {
	t.Helper()
	if _, err := ch.QueuePurge(queueName, false); err != nil {
		t.Fatalf("purge queue %s: %v", queueName, err)
	}
}
