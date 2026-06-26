package jobs

import (
	"io"
	"log"
	"testing"
)

func TestHandleTicketGeneratedSuccess(t *testing.T) {
	payload := []byte(`{"reservation_id":"r1","event_id":"e1","seat_id":"s1","user_id":"u1","payment_method":"card"}`)

	err := HandleTicketGenerated(log.New(io.Discard, "", 0), payload)
	if err != nil {
		t.Fatalf("HandleTicketGenerated() error = %v, want nil", err)
	}
}

func TestHandleTicketGeneratedMissingFieldIsPermanent(t *testing.T) {
	payload := []byte(`{"event_id":"e1","seat_id":"s1","user_id":"u1"}`)

	err := HandleTicketGenerated(log.New(io.Discard, "", 0), payload)
	if err == nil {
		t.Fatalf("HandleTicketGenerated() error = nil, want permanent error")
	}
	if !IsPermanentError(err) {
		t.Fatalf("HandleTicketGenerated() permanent = false, want true (err=%v)", err)
	}
}

func TestHandleTicketGeneratedTransientFailure(t *testing.T) {
	payload := []byte(`{"reservation_id":"r1","event_id":"e1","seat_id":"s1","user_id":"u1","payment_method":"simulate_transient"}`)

	err := HandleTicketGenerated(log.New(io.Discard, "", 0), payload)
	if err == nil {
		t.Fatalf("HandleTicketGenerated() error = nil, want transient error")
	}
	if IsPermanentError(err) {
		t.Fatalf("HandleTicketGenerated() permanent = true, want false")
	}
}
