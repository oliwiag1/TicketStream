package services

import "context"

type RealtimeService struct{}

func NewRealtimeService() *RealtimeService {
	return &RealtimeService{}
}

func (s *RealtimeService) BroadcastSeatUpdate(ctx context.Context, eventID string, payload []byte) error {
	_ = ctx
	_ = eventID
	_ = payload
	return nil
}
