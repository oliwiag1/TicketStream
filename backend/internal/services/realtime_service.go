package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/websocket"
)

const (
	seatStatusChangedType = "seat_status_changed"
	seatSnapshotType      = "snapshot"
	wsSequencePrefix      = "ws_sequence:"
)

var ErrEventNotFound = errors.New("event not found")

type SeatStatusChangedEvent struct {
	Type           string    `json:"type"`
	EventID        string    `json:"event_id"`
	SeatID         string    `json:"seat_id"`
	Status         string    `json:"status"`
	SequenceNumber int64     `json:"sequence_number"`
	ChangedAt      time.Time `json:"changed_at"`
}

type SeatSnapshotItem struct {
	SeatID    string    `json:"seat_id"`
	Row       string    `json:"row"`
	Number    int       `json:"number"`
	Status    string    `json:"status"`
	ChangedAt time.Time `json:"changed_at"`
}

type SeatSnapshotEvent struct {
	Type           string             `json:"type"`
	EventID        string             `json:"event_id"`
	SequenceNumber int64              `json:"sequence_number"`
	Seats          []SeatSnapshotItem `json:"seats"`
}

type RealtimeService struct {
	logger *log.Logger
	pgPool *pgxpool.Pool
	redis  *redis.Client

	mu        sync.RWMutex
	clients   map[string]map[*realtimeClient]struct{}
	sequences map[string]int64
}

type realtimeClient struct {
	conn   *websocket.Conn
	send   chan []byte
	closed chan struct{}
	once   sync.Once
}

func NewRealtimeService(logger *log.Logger, pgPool *pgxpool.Pool, redisClient *redis.Client) *RealtimeService {
	return &RealtimeService{
		logger:    logger,
		pgPool:    pgPool,
		redis:     redisClient,
		clients:   make(map[string]map[*realtimeClient]struct{}),
		sequences: make(map[string]int64),
	}
}

func (s *RealtimeService) EventExists(ctx context.Context, eventID string) (bool, error) {
	if s.pgPool == nil {
		return false, nil
	}

	var exists bool
	err := s.pgPool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM events WHERE id = $1)`, eventID).Scan(&exists)
	return exists, err
}

func (s *RealtimeService) ServeConnection(ctx context.Context, eventID string, conn *websocket.Conn) error {
	client := &realtimeClient{
		conn:   conn,
		send:   make(chan []byte, 16),
		closed: make(chan struct{}),
	}

	s.registerClient(eventID, client)
	defer s.unregisterClient(eventID, client)

	snapshot, err := s.BuildSnapshot(ctx, eventID)
	if err != nil {
		return err
	}

	writerDone := make(chan error, 1)
	go func() {
		writerDone <- client.writeLoop()
	}()

	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	if err := client.enqueue(payload); err != nil {
		return err
	}

	readerDone := make(chan error, 1)
	go func() {
		readerDone <- client.readLoop()
	}()

	select {
	case err := <-writerDone:
		return normalizeConnectionError(err)
	case err := <-readerDone:
		return normalizeConnectionError(err)
	case <-ctx.Done():
		client.close()
		return nil
	}
}

func (s *RealtimeService) BuildSnapshot(ctx context.Context, eventID string) (SeatSnapshotEvent, error) {
	exists, err := s.EventExists(ctx, eventID)
	if err != nil {
		return SeatSnapshotEvent{}, err
	}
	if !exists {
		return SeatSnapshotEvent{}, ErrEventNotFound
	}

	rows, err := s.pgPool.Query(
		ctx,
		`SELECT id, seat_row, seat_number, status, updated_at
		 FROM seats
		 WHERE event_id = $1
		 ORDER BY seat_row, seat_number`,
		eventID,
	)
	if err != nil {
		return SeatSnapshotEvent{}, err
	}
	defer rows.Close()

	seats := make([]SeatSnapshotItem, 0)
	for rows.Next() {
		var seat SeatSnapshotItem
		if err := rows.Scan(&seat.SeatID, &seat.Row, &seat.Number, &seat.Status, &seat.ChangedAt); err != nil {
			return SeatSnapshotEvent{}, err
		}
		seats = append(seats, seat)
	}
	if err := rows.Err(); err != nil {
		return SeatSnapshotEvent{}, err
	}

	sequenceNumber, err := s.currentSequenceNumber(ctx, eventID)
	if err != nil {
		return SeatSnapshotEvent{}, err
	}

	return SeatSnapshotEvent{
		Type:           seatSnapshotType,
		EventID:        eventID,
		SequenceNumber: sequenceNumber,
		Seats:          seats,
	}, nil
}

func (s *RealtimeService) BroadcastSeatUpdate(ctx context.Context, eventID, seatID, status string, changedAt time.Time) error {
	sequenceNumber, err := s.nextSequenceNumber(ctx, eventID)
	if err != nil {
		return err
	}

	event := SeatStatusChangedEvent{
		Type:           seatStatusChangedType,
		EventID:        eventID,
		SeatID:         seatID,
		Status:         status,
		SequenceNumber: sequenceNumber,
		ChangedAt:      changedAt.UTC(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	s.broadcast(eventID, payload)
	return nil
}

func (s *RealtimeService) registerClient(eventID string, client *realtimeClient) {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventClients := s.clients[eventID]
	if eventClients == nil {
		eventClients = make(map[*realtimeClient]struct{})
		s.clients[eventID] = eventClients
	}
	eventClients[client] = struct{}{}
}

func (s *RealtimeService) unregisterClient(eventID string, client *realtimeClient) {
	client.close()

	s.mu.Lock()
	defer s.mu.Unlock()

	eventClients := s.clients[eventID]
	if eventClients == nil {
		return
	}
	delete(eventClients, client)
	if len(eventClients) == 0 {
		delete(s.clients, eventID)
	}
}

func (s *RealtimeService) broadcast(eventID string, payload []byte) {
	s.mu.RLock()
	eventClients := s.clients[eventID]
	clients := make([]*realtimeClient, 0, len(eventClients))
	for client := range eventClients {
		clients = append(clients, client)
	}
	s.mu.RUnlock()

	for _, client := range clients {
		if err := client.enqueue(payload); err != nil {
			s.unregisterClient(eventID, client)
		}
	}
}

func (s *RealtimeService) nextSequenceNumber(ctx context.Context, eventID string) (int64, error) {
	if s.redis == nil {
		return s.incrementLocalSequence(eventID), nil
	}

	nextValue, err := s.redis.Incr(ctx, wsSequenceKey(eventID)).Result()
	if err != nil {
		return 0, err
	}

	s.mu.Lock()
	s.sequences[eventID] = nextValue
	s.mu.Unlock()

	return nextValue, nil
}

func (s *RealtimeService) currentSequenceNumber(ctx context.Context, eventID string) (int64, error) {
	if s.redis == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return s.sequences[eventID], nil
	}

	rawValue, err := s.redis.Get(ctx, wsSequenceKey(eventID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}

	parsed, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func (s *RealtimeService) incrementLocalSequence(eventID string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sequences[eventID]++
	return s.sequences[eventID]
}

func wsSequenceKey(eventID string) string {
	return wsSequencePrefix + eventID
}

func (c *realtimeClient) enqueue(payload []byte) error {
	select {
	case <-c.closed:
		return io.EOF
	default:
	}

	select {
	case c.send <- payload:
		return nil
	default:
		c.close()
		return io.ErrClosedPipe
	}
}

func (c *realtimeClient) close() {
	c.once.Do(func() {
		close(c.closed)
		close(c.send)
		_ = c.conn.Close()
	})
}

func (c *realtimeClient) writeLoop() error {
	for payload := range c.send {
		if err := websocket.Message.Send(c.conn, string(payload)); err != nil {
			c.close()
			return err
		}
	}
	return nil
}

func (c *realtimeClient) readLoop() error {
	for {
		var message string
		if err := websocket.Message.Receive(c.conn, &message); err != nil {
			c.close()
			return err
		}
	}
}

func normalizeConnectionError(err error) error {
	if err == nil || errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
		return nil
	}
	return err
}
