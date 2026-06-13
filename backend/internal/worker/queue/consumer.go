package queue

import (
	"log"

	"ticketstream/backend/internal/worker/jobs"

	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	ch     *amqp091.Channel
	logger *log.Logger
}

func NewConsumer(ch *amqp091.Channel, logger *log.Logger) *Consumer {
	return &Consumer{ch: ch, logger: logger}
}

func (c *Consumer) Start() error {
	if err := c.ch.ExchangeDeclare("ticketstream.events", "topic", true, false, false, false, nil); err != nil {
		return err
	}

	queue, err := c.ch.QueueDeclare("ticketstream.ticket.jobs", true, false, false, false, nil)
	if err != nil {
		return err
	}

	if err := c.ch.QueueBind(queue.Name, "ticket.generated", "ticketstream.events", false, nil); err != nil {
		return err
	}

	messages, err := c.ch.Consume(queue.Name, "ticketstream-worker", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for msg := range messages {
		if err := jobs.HandleTicketGenerated(c.logger, msg.Body); err != nil {
			c.logger.Printf("job failed: %v", err)
			msg.Nack(false, true)
			continue
		}
		msg.Ack(false)
	}

	return nil
}
