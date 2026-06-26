package queue

import (
	"errors"
	"testing"

	"ticketstream/backend/internal/worker/jobs"

	"github.com/rabbitmq/amqp091-go"
)

func TestRetryCountFromHeaders(t *testing.T) {
	headers := amqp091.Table{RetryCountHeader: int32(2)}

	got := retryCountFromHeaders(headers)
	if got != 2 {
		t.Fatalf("retryCountFromHeaders() = %d, want 2", got)
	}
}

func TestDecideMessageOutcome(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		retryCount int
		want       messageOutcome
	}{
		{name: "ack on success", err: nil, retryCount: 0, want: outcomeAck},
		{name: "permanent error to dlq", err: jobs.NewPermanentError("bad payload"), retryCount: 0, want: outcomeDLQ},
		{name: "transient error retries", err: errors.New("temporary"), retryCount: 1, want: outcomeRetry},
		{name: "transient error hits dlq after limit", err: errors.New("temporary"), retryCount: MaxRetryAttempts, want: outcomeDLQ},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decideMessageOutcome(tt.err, tt.retryCount)
			if got != tt.want {
				t.Fatalf("decideMessageOutcome() = %v, want %v", got, tt.want)
			}
		})
	}
}
