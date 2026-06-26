//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ticketstream/backend/internal/config"
	"ticketstream/backend/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

func TestIntegrationWebSocketSnapshotReconnectAndSeatBroadcasts(t *testing.T) {
	ctx := context.Background()
	pgPool, redisClient := integrationDependencies(t)
	defer pgPool.Close()
	defer redisClient.Close()

	eventID := uuid.NewString()
	seatOneID := uuid.NewString()
	seatTwoID := uuid.NewString()

	_, err := pgPool.Exec(
		ctx,
		`INSERT INTO events (id, title, description, starts_at) VALUES ($1, $2, $3, $4)`,
		eventID,
		"Realtime Concert",
		"ws test",
		time.Now().UTC().Add(2*time.Hour),
	)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}
	_, err = pgPool.Exec(
		ctx,
		`INSERT INTO seats (id, event_id, seat_row, seat_number, status) VALUES
		 ($1, $3, 'A', 1, 'available'),
		 ($2, $3, 'A', 2, 'available')`,
		seatOneID,
		seatTwoID,
		eventID,
	)
	if err != nil {
		t.Fatalf("insert seats: %v", err)
	}
	if err := redisClient.Del(ctx, "ws_sequence:"+eventID).Err(); err != nil {
		t.Fatalf("reset websocket sequence: %v", err)
	}

	realtimeService := services.NewRealtimeService(log.New(io.Discard, "", 0), pgPool, redisClient)
	wsHandler := NewWSHandler(log.New(io.Discard, "", 0), realtimeService)
	reservationHandler := NewReservationHandler(config.Config{SeatLockTTLSeconds: 600}, log.New(io.Discard, "", 0), pgPool, redisClient, realtimeService)
	paymentHandler := NewPaymentHandler(log.New(io.Discard, "", 0), pgPool, redisClient, realtimeService)

	e := echo.New()
	e.GET("/ws/events/:eventId", wsHandler.Connect)
	server := httptest.NewServer(e)
	defer server.Close()

	clientOne := mustDialEventSocket(t, server.URL, eventID)
	defer clientOne.Close()
	clientTwo := mustDialEventSocket(t, server.URL, eventID)
	defer clientTwo.Close()

	assertSnapshot(t, readSnapshot(t, clientOne), eventID, 0, map[string]string{
		seatOneID: "available",
		seatTwoID: "available",
	})
	assertSnapshot(t, readSnapshot(t, clientTwo), eventID, 0, map[string]string{
		seatOneID: "available",
		seatTwoID: "available",
	})

	subject := "ws-user-" + uuid.NewString()

	firstReservation := reserveSeatForTest(t, reservationHandler, eventID, seatOneID, subject)
	assertSeatUpdate(t, readSeatUpdate(t, clientOne), eventID, seatOneID, "locked", 1)
	assertSeatUpdate(t, readSeatUpdate(t, clientTwo), eventID, seatOneID, "locked", 1)

	cancelReservationForTest(t, reservationHandler, firstReservation.ReservationID, subject)
	assertSeatUpdate(t, readSeatUpdate(t, clientOne), eventID, seatOneID, "available", 2)
	assertSeatUpdate(t, readSeatUpdate(t, clientTwo), eventID, seatOneID, "available", 2)

	secondReservation := reserveSeatForTest(t, reservationHandler, eventID, seatTwoID, subject)
	assertSeatUpdate(t, readSeatUpdate(t, clientOne), eventID, seatTwoID, "locked", 3)
	assertSeatUpdate(t, readSeatUpdate(t, clientTwo), eventID, seatTwoID, "locked", 3)

	payReservationForTest(t, paymentHandler, secondReservation.ReservationID, subject)
	assertSeatUpdate(t, readSeatUpdate(t, clientOne), eventID, seatTwoID, "sold", 4)
	assertSeatUpdate(t, readSeatUpdate(t, clientTwo), eventID, seatTwoID, "sold", 4)

	if err := clientTwo.Close(); err != nil {
		t.Fatalf("close second websocket client: %v", err)
	}

	reconnectedClient := mustDialEventSocket(t, server.URL, eventID)
	defer reconnectedClient.Close()

	assertSnapshot(t, readSnapshot(t, reconnectedClient), eventID, 4, map[string]string{
		seatOneID: "available",
		seatTwoID: "sold",
	})
}

func mustDialEventSocket(t *testing.T, serverURL, eventID string) *websocket.Conn {
	t.Helper()

	socketURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/ws/events/" + eventID
	conn, err := websocket.Dial(socketURL, "", "http://localhost/")
	if err != nil {
		t.Fatalf("dial websocket %s: %v", socketURL, err)
	}
	return conn
}

func readSnapshot(t *testing.T, conn *websocket.Conn) services.SeatSnapshotEvent {
	t.Helper()

	var rawMessage string
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("set websocket deadline: %v", err)
	}
	if err := websocket.Message.Receive(conn, &rawMessage); err != nil {
		t.Fatalf("receive websocket snapshot: %v", err)
	}

	var snapshot services.SeatSnapshotEvent
	if err := json.Unmarshal([]byte(rawMessage), &snapshot); err != nil {
		t.Fatalf("decode snapshot payload: %v", err)
	}
	return snapshot
}

func readSeatUpdate(t *testing.T, conn *websocket.Conn) services.SeatStatusChangedEvent {
	t.Helper()

	var rawMessage string
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("set websocket deadline: %v", err)
	}
	if err := websocket.Message.Receive(conn, &rawMessage); err != nil {
		t.Fatalf("receive websocket event: %v", err)
	}

	var event services.SeatStatusChangedEvent
	if err := json.Unmarshal([]byte(rawMessage), &event); err != nil {
		t.Fatalf("decode websocket event: %v", err)
	}
	return event
}

func assertSnapshot(t *testing.T, snapshot services.SeatSnapshotEvent, eventID string, sequenceNumber int64, expectedStatuses map[string]string) {
	t.Helper()

	if snapshot.Type != "snapshot" {
		t.Fatalf("snapshot type = %s, want snapshot", snapshot.Type)
	}
	if snapshot.EventID != eventID {
		t.Fatalf("snapshot event_id = %s, want %s", snapshot.EventID, eventID)
	}
	if snapshot.SequenceNumber != sequenceNumber {
		t.Fatalf("snapshot sequence_number = %d, want %d", snapshot.SequenceNumber, sequenceNumber)
	}
	if len(snapshot.Seats) != len(expectedStatuses) {
		t.Fatalf("snapshot seats len = %d, want %d", len(snapshot.Seats), len(expectedStatuses))
	}

	for _, seat := range snapshot.Seats {
		expectedStatus, ok := expectedStatuses[seat.SeatID]
		if !ok {
			t.Fatalf("unexpected snapshot seat_id = %s", seat.SeatID)
		}
		if seat.Status != expectedStatus {
			t.Fatalf("snapshot status for seat %s = %s, want %s", seat.SeatID, seat.Status, expectedStatus)
		}
	}
}

func assertSeatUpdate(t *testing.T, event services.SeatStatusChangedEvent, eventID, seatID, status string, sequenceNumber int64) {
	t.Helper()

	if event.Type != "seat_status_changed" {
		t.Fatalf("event type = %s, want seat_status_changed", event.Type)
	}
	if event.EventID != eventID {
		t.Fatalf("event event_id = %s, want %s", event.EventID, eventID)
	}
	if event.SeatID != seatID {
		t.Fatalf("event seat_id = %s, want %s", event.SeatID, seatID)
	}
	if event.Status != status {
		t.Fatalf("event status = %s, want %s", event.Status, status)
	}
	if event.SequenceNumber != sequenceNumber {
		t.Fatalf("event sequence_number = %d, want %d", event.SequenceNumber, sequenceNumber)
	}
}

func reserveSeatForTest(t *testing.T, reservationHandler *ReservationHandler, eventID, seatID, subject string) reserveResponsePayload {
	t.Helper()

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

	var payload reserveResponsePayload
	if err := json.Unmarshal(reserveRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode reserve response: %v", err)
	}
	return payload
}

func cancelReservationForTest(t *testing.T, reservationHandler *ReservationHandler, reservationID, subject string) {
	t.Helper()

	cancelCtx, cancelRec := newAuthedContext(t, subject, "DELETE", "/reservations/"+reservationID, nil)
	cancelCtx.SetParamNames("reservationId")
	cancelCtx.SetParamValues(reservationID)

	if err := reservationHandler.Cancel(cancelCtx); err != nil {
		t.Fatalf("cancel handler returned error: %v", err)
	}
	if cancelRec.Code != 200 {
		t.Fatalf("cancel status = %d, want 200, body=%s", cancelRec.Code, cancelRec.Body.String())
	}
}

func payReservationForTest(t *testing.T, paymentHandler *PaymentHandler, reservationID, subject string) {
	t.Helper()

	payBody := []byte(`{"reservation_id":"` + reservationID + `","payment_method":"card"}`)
	payCtx, payRec := newAuthedContext(t, subject, "POST", "/pay", payBody)
	payCtx.Request().Header.Set("Idempotency-Key", uuid.NewString())

	if err := paymentHandler.Pay(payCtx); err != nil {
		t.Fatalf("payment handler returned error: %v", err)
	}
	if payRec.Code != 200 {
		t.Fatalf("payment status = %d, want 200, body=%s", payRec.Code, payRec.Body.String())
	}
}

type reserveResponsePayload struct {
	ReservationID string `json:"reservation_id"`
}
