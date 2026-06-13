package services

import "context"

type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

func (s *PaymentService) PayReservation(ctx context.Context, reservationID, idempotencyKey string) error {
	_ = ctx
	_ = reservationID
	_ = idempotencyKey
	return nil
}
