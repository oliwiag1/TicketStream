package services

import "context"

type ReservationService struct{}

func NewReservationService() *ReservationService {
	return &ReservationService{}
}

func (s *ReservationService) ReserveSeat(ctx context.Context, userID, eventID, seatID string) error {
	_ = ctx
	_ = userID
	_ = eventID
	_ = seatID
	return nil
}
