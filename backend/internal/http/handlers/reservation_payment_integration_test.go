//go:build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"ticketstream/backend/internal/auth"
	"ticketstream/backend/internal/config"
	httpmiddleware "ticketstream/backend/internal/http/middleware"
	"ticketstream/backend/pkg/db"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

func TestIntegrationConcurrentReservationSingleWinner(t *testing.T) {
	ctx := context.Background()
	pgPool, redisClient := integrationDependencies(t)
	defer pgPool.Close()
	defer redisClient.Close()

	eventID := uuid.NewString()
	seatID := uuid.NewString()
	_, err := pgPool.Exec(
		ctx,
		`INSERT INTO events (id, title, description, starts_at) VALUES ($1, $2, $3, $4)`,
		eventID,
		"Concurrent Reservation Concert",
		"anti-oversell",
		time.Now().UTC().Add(2*time.Hour),
	)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}
	_, err = pgPool.Exec(
		ctx,
		`INSERT INTO seats (id, event_id, seat_row, seat_number, status) VALUES ($1, $2, 'A', 10, 'available')`,
		seatID,
		eventID,
	)
	if err != nil {
		t.Fatalf("insert seat: %v", err)
	}

	reservationHandler := NewReservationHandler(config.Config{SeatLockTTLSeconds: 600}, log.New(io.Discard, "", 0), pgPool, redisClient)

	const attempts = 8
	statusCodes := make([]int, attempts)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(attempts)

	for i := 0; i < attempts; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start

			subject := "concurrent-user-" + uuid.NewString()
			reserveBody := []byte(`{"seat_id":"` + seatID + `"}`)
			reserveCtx, reserveRec := newAuthedContext(t, subject, "POST", "/events/"+eventID+"/reserve", reserveBody)
			reserveCtx.SetParamNames("eventId")
			reserveCtx.SetParamValues(eventID)

			if err := reservationHandler.Reserve(reserveCtx); err != nil {
				statusCodes[i] = http.StatusInternalServerError
				return
			}
			statusCodes[i] = reserveRec.Code
		}()
	}

	close(start)
	wg.Wait()

	successCount := 0
	conflictCount := 0
	for _, code := range statusCodes {
		switch code {
		case http.StatusAccepted:
			successCount++
		case http.StatusConflict:
			conflictCount++
		default:
			t.Fatalf("unexpected status code from concurrent reserve: %d", code)
		}
	}

	if successCount != 1 {
		t.Fatalf("successCount = %d, want 1 (codes=%v)", successCount, statusCodes)
	}
	if conflictCount != attempts-1 {
		t.Fatalf("conflictCount = %d, want %d (codes=%v)", conflictCount, attempts-1, statusCodes)
	}

	var seatStatus string
	if err := pgPool.QueryRow(ctx, `SELECT status FROM seats WHERE id = $1`, seatID).Scan(&seatStatus); err != nil {
		t.Fatalf("query seat status after concurrent reserve: %v", err)
	}
	if seatStatus != "locked" {
		t.Fatalf("seat status = %s, want locked", seatStatus)
	}

	var reservationCount int
	if err := pgPool.QueryRow(ctx, `SELECT COUNT(*) FROM reservations WHERE seat_id = $1 AND status = 'locked'`, seatID).Scan(&reservationCount); err != nil {
		t.Fatalf("count locked reservations: %v", err)
	}
	if reservationCount != 1 {
		t.Fatalf("locked reservation count = %d, want 1", reservationCount)
	}
}

func TestIntegrationReservationAndPaymentFlow(t *testing.T) {
	ctx := context.Background()
	pgPool, redisClient := integrationDependencies(t)
	defer pgPool.Close()
	defer redisClient.Close()

	eventID := uuid.NewString()
	seatID := uuid.NewString()
	_, err := pgPool.Exec(
		ctx,
		`INSERT INTO events (id, title, description, starts_at) VALUES ($1, $2, $3, $4)`,
		eventID,
		"Integration Concert",
		"flow test",
		time.Now().UTC().Add(2*time.Hour),
	)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}
	_, err = pgPool.Exec(
		ctx,
		`INSERT INTO seats (id, event_id, seat_row, seat_number, status) VALUES ($1, $2, 'A', 1, 'available')`,
		seatID,
		eventID,
	)
	if err != nil {
		t.Fatalf("insert seat: %v", err)
	}

	reservationHandler := NewReservationHandler(config.Config{SeatLockTTLSeconds: 600}, log.New(io.Discard, "", 0), pgPool, redisClient)
	paymentHandler := NewPaymentHandler(log.New(io.Discard, "", 0), pgPool, redisClient)
	subject := "integration-user-" + uuid.NewString()

	reserveBody := []byte(`{"seat_id":"` + seatID + `"}`)
	reserveCtx, reserveRec := newAuthedContext(t, subject, "POST", "/events/"+eventID+"/reserve", reserveBody)
	reserveCtx.SetParamNames("eventId")
	reserveCtx.SetParamValues(eventID)

	if err := reservationHandler.Reserve(reserveCtx); err != nil {
		t.Fatalf("reserve handler returned error: %v", err)
	}
	if reserveRec.Code != 202 {
		t.Fatalf("reserve status = %d, want 202, body=%s", reserveRec.Code, reserveRec.Body.String())
	}

	var reserveResp struct {
		ReservationID string `json:"reservation_id"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal(reserveRec.Body.Bytes(), &reserveResp); err != nil {
		t.Fatalf("decode reserve response: %v", err)
	}
	if reserveResp.ReservationID == "" || reserveResp.Status != "locked" {
		t.Fatalf("unexpected reserve response: %+v", reserveResp)
	}

	var seatStatus string
	if err := pgPool.QueryRow(ctx, `SELECT status FROM seats WHERE id = $1`, seatID).Scan(&seatStatus); err != nil {
		t.Fatalf("query seat status after reserve: %v", err)
	}
	if seatStatus != "locked" {
		t.Fatalf("seat status after reserve = %s, want locked", seatStatus)
	}

	idempotencyKey := uuid.NewString()
	payBody := []byte(`{"reservation_id":"` + reserveResp.ReservationID + `","payment_method":"card"}`)
	payCtx, payRec := newAuthedContext(t, subject, "POST", "/pay", payBody)
	payCtx.Request().Header.Set("Idempotency-Key", idempotencyKey)

	if err := paymentHandler.Pay(payCtx); err != nil {
		t.Fatalf("payment handler returned error: %v", err)
	}
	if payRec.Code != 200 {
		t.Fatalf("payment status = %d, want 200, body=%s", payRec.Code, payRec.Body.String())
	}

	payRetryCtx, payRetryRec := newAuthedContext(t, subject, "POST", "/pay", payBody)
	payRetryCtx.Request().Header.Set("Idempotency-Key", idempotencyKey)
	if err := paymentHandler.Pay(payRetryCtx); err != nil {
		t.Fatalf("payment retry handler returned error: %v", err)
	}
	if payRetryRec.Code != 200 {
		t.Fatalf("payment retry status = %d, want 200, body=%s", payRetryRec.Code, payRetryRec.Body.String())
	}

	var reservationStatus string
	if err := pgPool.QueryRow(ctx, `SELECT status FROM reservations WHERE id = $1`, reserveResp.ReservationID).Scan(&reservationStatus); err != nil {
		t.Fatalf("query reservation status: %v", err)
	}
	if reservationStatus != "sold" {
		t.Fatalf("reservation status = %s, want sold", reservationStatus)
	}

	if err := pgPool.QueryRow(ctx, `SELECT status FROM seats WHERE id = $1`, seatID).Scan(&seatStatus); err != nil {
		t.Fatalf("query seat status after payment: %v", err)
	}
	if seatStatus != "sold" {
		t.Fatalf("seat status after payment = %s, want sold", seatStatus)
	}

	var outboxEventType string
	var publishedAt *time.Time
	if err := pgPool.QueryRow(
		ctx,
		`SELECT event_type, published_at
		 FROM outbox_events
		 WHERE aggregate_id = $1
		 ORDER BY created_at DESC
		 LIMIT 1`,
		reserveResp.ReservationID,
	).Scan(&outboxEventType, &publishedAt); err != nil {
		t.Fatalf("query outbox event: %v", err)
	}
	if outboxEventType != "payment.succeeded" {
		t.Fatalf("outbox event_type = %s, want payment.succeeded", outboxEventType)
	}
	if publishedAt != nil {
		t.Fatalf("outbox published_at = %v, want nil before publisher runs", *publishedAt)
	}
}

func integrationDependencies(t *testing.T) (*pgxpool.Pool, *redis.Client) {
	t.Helper()
	ctx := context.Background()

	postgresURL := os.Getenv("TEST_POSTGRES_URL")
	if postgresURL == "" {
		postgresURL = "postgres://ticketstream:ticketstream@localhost:5433/ticketstream?sslmode=disable"
	}

	pgPool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		t.Skipf("skipping integration test: cannot create postgres pool: %v", err)
	}
	if err := pgPool.Ping(ctx); err != nil {
		pgPool.Close()
		t.Skipf("skipping integration test: postgres unavailable: %v", err)
	}

	migrationsDir := os.Getenv("TEST_MIGRATIONS_DIR")
	if migrationsDir == "" {
		wd, wdErr := os.Getwd()
		if wdErr != nil {
			pgPool.Close()
			t.Skipf("skipping integration test: cannot resolve working dir: %v", wdErr)
		}
		migrationsDir = filepath.Clean(filepath.Join(wd, "..", "..", "..", "migrations"))
	}

	if err := db.RunMigrations(ctx, pgPool, migrationsDir, log.New(io.Discard, "", 0)); err != nil {
		pgPool.Close()
		t.Skipf("skipping integration test: migrations failed: %v", err)
	}

	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if parsed, err := redis.ParseURL(redisURL); err == nil {
		redisClient = redis.NewClient(parsed)
	}
	if err := redisClient.Ping(ctx).Err(); err != nil {
		pgPool.Close()
		redisClient.Close()
		t.Skipf("skipping integration test: redis unavailable: %v", err)
	}

	return pgPool, redisClient
}

func newAuthedContext(t *testing.T, subject, method, path string, body []byte) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	claims := &auth.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: subject},
		Subject:          subject,
		Email:            subject + "@example.local",
		RealmAccess:      auth.RealmAccessClaims{Roles: []string{"user"}},
	}
	c.Set(httpmiddleware.ClaimsContextKey, claims)
	return c, rec
}
