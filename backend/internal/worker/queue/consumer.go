package queue

import (
	"context"
	"fmt"
	"log"

	"ticketstream/backend/internal/worker/jobs"

	"github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName      = "ticketstream.events"
	TicketQueueName   = "ticketstream.ticket.jobs"
	TicketDLQName     = "ticketstream.ticket.jobs.dlq"
	TicketRoutingKey  = "ticket.generated"
	TicketDLQRouteKey = "ticket.generated.dlq"
	MaxRetryAttempts  = 3
	RetryCountHeader  = "x-retry-count"
)

type Consumer struct {
	ch     *amqp091.Channel
	logger *log.Logger
}

type messageOutcome int

const (
	outcomeAck messageOutcome = iota
	outcomeRetry
	outcomeDLQ
)

func NewConsumer(ch *amqp091.Channel, logger *log.Logger) *Consumer {
	return &Consumer{ch: ch, logger: logger}
}

func (c *Consumer) Start() error {
	if err := c.ch.ExchangeDeclare(ExchangeName, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	queue, err := c.ch.QueueDeclare(TicketQueueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	dlq, err := c.ch.QueueDeclare(TicketDLQName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	if err := c.ch.QueueBind(queue.Name, TicketRoutingKey, ExchangeName, false, nil); err != nil {
		return err
	}

	if err := c.ch.QueueBind(dlq.Name, TicketDLQRouteKey, ExchangeName, false, nil); err != nil {
		return err
	}

	messages, err := c.ch.Consume(queue.Name, "ticketstream-worker", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for msg := range messages {
		if err := c.handleMessage(context.Background(), msg); err != nil {
			c.logger.Printf("worker message handling failed: %v", err)
			continue
		}
	}

	return nil
}

func (c *Consumer) handleMessage(ctx context.Context, msg amqp091.Delivery) error {
	err := jobs.HandleTicketGenerated(c.logger, msg.Body)
	retryCount := retryCountFromHeaders(msg.Headers)

	switch decideMessageOutcome(err, retryCount) {
	case outcomeAck:
		return msg.Ack(false)
	case outcomeDLQ:
		return c.publishToDLQ(ctx, msg, retryCount, err)
	case outcomeRetry:
		if err := c.requeueWithRetry(ctx, msg, retryCount+1); err != nil {
			return err
		}
		return msg.Ack(false)
	default:
		return fmt.Errorf("unsupported message outcome for retry_count=%d", retryCount)
	}
}

func (c *Consumer) requeueWithRetry(ctx context.Context, msg amqp091.Delivery, retryCount int) error {
	headers := cloneHeaders(msg.Headers)
	headers[RetryCountHeader] = retryCount
	headers["x-last-error"] = "transient_failure"

	if err := c.ch.PublishWithContext(
		ctx,
		ExchangeName,
		TicketRoutingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  contentTypeOrDefault(msg.ContentType),
			DeliveryMode: amqp091.Persistent,
			Body:         msg.Body,
			Headers:      headers,
		},
	); err != nil {
		return fmt.Errorf("requeue ticket job: %w", err)
	}

	return nil
}

func (c *Consumer) publishToDLQ(ctx context.Context, msg amqp091.Delivery, retryCount int, jobErr error) error {
	headers := cloneHeaders(msg.Headers)
	headers[RetryCountHeader] = retryCount
	headers["x-error"] = jobErr.Error()

	if err := c.ch.PublishWithContext(
		ctx,
		ExchangeName,
		TicketDLQRouteKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  contentTypeOrDefault(msg.ContentType),
			DeliveryMode: amqp091.Persistent,
			Body:         msg.Body,
			Headers:      headers,
		},
	); err != nil {
		return fmt.Errorf("publish ticket job to dlq: %w", err)
	}

	return msg.Ack(false)
}

func retryCountFromHeaders(headers amqp091.Table) int {
	if headers == nil {
		return 0
	}

	value, ok := headers[RetryCountHeader]
	if !ok {
		return 0
	}

	switch typed := value.(type) {
	case int:
		return typed
	case int8:
		return int(typed)
	case int16:
		return int(typed)
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	default:
		return 0
	}
}

func cloneHeaders(headers amqp091.Table) amqp091.Table {
	if headers == nil {
		return amqp091.Table{}
	}

	cloned := make(amqp091.Table, len(headers))
	for key, value := range headers {
		cloned[key] = value
	}
	return cloned
}

func contentTypeOrDefault(contentType string) string {
	if contentType == "" {
		return "application/json"
	}
	return contentType
}

func decideMessageOutcome(jobErr error, retryCount int) messageOutcome {
	if jobErr == nil {
		return outcomeAck
	}

	if jobs.IsPermanentError(jobErr) {
		return outcomeDLQ
	}

	if retryCount >= MaxRetryAttempts {
		return outcomeDLQ
	}

	return outcomeRetry
}
