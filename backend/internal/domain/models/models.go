package models

import "time"

type SeatStatus string

const (
	SeatStatusAvailable SeatStatus = "available"
	SeatStatusLocked    SeatStatus = "locked"
	SeatStatusSold      SeatStatus = "sold"
)

type User struct {
	ID        string
	Email     string
	Password  string
	CreatedAt time.Time
}

type Event struct {
	ID          string
	Title       string
	StartsAt    time.Time
	CreatedAt   time.Time
	Description string
}

type Seat struct {
	ID        string
	EventID   string
	Row       string
	Number    int
	Status    SeatStatus
	UpdatedAt time.Time
}

type Reservation struct {
	ID        string
	UserID    string
	EventID   string
	SeatID    string
	Status    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type PaymentAttempt struct {
	ID             string
	ReservationID  string
	IdempotencyKey string
	Status         string
	CreatedAt      time.Time
}

type OutboxEvent struct {
	ID          string
	AggregateID string
	EventType   string
	PayloadJSON []byte
	CreatedAt   time.Time
	PublishedAt *time.Time
}
