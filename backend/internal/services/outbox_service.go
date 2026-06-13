package services

import "context"

type OutboxService struct{}

func NewOutboxService() *OutboxService {
	return &OutboxService{}
}

func (s *OutboxService) PublishPending(ctx context.Context) error {
	_ = ctx
	return nil
}
