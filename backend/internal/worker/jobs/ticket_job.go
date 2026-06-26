package jobs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type TicketGeneratedPayload struct {
	ReservationID string `json:"reservation_id"`
	EventID       string `json:"event_id"`
	SeatID        string `json:"seat_id"`
	UserID        string `json:"user_id"`
	PaymentMethod string `json:"payment_method"`
}

type PermanentError struct {
	err error
}

func (e *PermanentError) Error() string {
	return e.err.Error()
}

func (e *PermanentError) Unwrap() error {
	return e.err
}

func NewPermanentError(message string) error {
	return &PermanentError{err: errors.New(message)}
}

func IsPermanentError(err error) bool {
	var permanentErr *PermanentError
	return errors.As(err, &permanentErr)
}

func HandleTicketGenerated(logger *log.Logger, payload []byte) error {
	var jobPayload TicketGeneratedPayload
	if err := json.Unmarshal(payload, &jobPayload); err != nil {
		return &PermanentError{err: fmt.Errorf("invalid ticket payload json: %w", err)}
	}

	if jobPayload.ReservationID == "" {
		return NewPermanentError("missing reservation_id")
	}
	if jobPayload.EventID == "" {
		return NewPermanentError("missing event_id")
	}
	if jobPayload.SeatID == "" {
		return NewPermanentError("missing seat_id")
	}
	if jobPayload.UserID == "" {
		return NewPermanentError("missing user_id")
	}
	if jobPayload.PaymentMethod == "simulate_transient" {
		return fmt.Errorf("temporary ticket provider outage")
	}

	logger.Printf(
		"ticket prepared reservation_id=%s event_id=%s seat_id=%s user_id=%s payment_method=%s",
		jobPayload.ReservationID,
		jobPayload.EventID,
		jobPayload.SeatID,
		jobPayload.UserID,
		jobPayload.PaymentMethod,
	)
	logger.Printf("email scheduled reservation_id=%s", jobPayload.ReservationID)
	return nil
}
