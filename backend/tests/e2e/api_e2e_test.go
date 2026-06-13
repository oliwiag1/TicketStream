//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestE2EPublicAPIAndAuthGuard(t *testing.T) {
	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	healthResp := mustRequest(t, client, "GET", baseURL+"/health", nil, nil)
	if healthResp.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d, want %d", healthResp.StatusCode, http.StatusOK)
	}
	defer healthResp.Body.Close()

	var healthBody map[string]string
	if err := json.NewDecoder(healthResp.Body).Decode(&healthBody); err != nil {
		t.Fatalf("decode /health response: %v", err)
	}
	if healthBody["status"] != "ok" {
		t.Fatalf("/health status payload = %q, want ok", healthBody["status"])
	}

	eventsResp := mustRequest(t, client, "GET", baseURL+"/events", nil, nil)
	if eventsResp.StatusCode != http.StatusOK {
		t.Fatalf("events status = %d, want %d", eventsResp.StatusCode, http.StatusOK)
	}
	defer eventsResp.Body.Close()

	var eventsBody struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(eventsResp.Body).Decode(&eventsBody); err != nil {
		t.Fatalf("decode /events response: %v", err)
	}
	if len(eventsBody.Items) == 0 || eventsBody.Items[0].ID == "" {
		t.Fatalf("/events should return at least one event")
	}

	seatsResp := mustRequest(t, client, "GET", baseURL+"/events/"+eventsBody.Items[0].ID+"/seats", nil, nil)
	if seatsResp.StatusCode != http.StatusOK {
		t.Fatalf("seats status = %d, want %d", seatsResp.StatusCode, http.StatusOK)
	}
	defer seatsResp.Body.Close()

	var seatsBody struct {
		Seats []map[string]any `json:"seats"`
	}
	if err := json.NewDecoder(seatsResp.Body).Decode(&seatsBody); err != nil {
		t.Fatalf("decode /events/:id/seats response: %v", err)
	}
	if len(seatsBody.Seats) == 0 {
		t.Fatalf("/events/:id/seats should return seats")
	}

	reservePayload := []byte(`{"seat_id":"some-seat"}`)
	reserveResp := mustRequest(t, client, "POST", baseURL+"/events/"+eventsBody.Items[0].ID+"/reserve", reservePayload, map[string]string{"Content-Type": "application/json"})
	if reserveResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("reserve without token status = %d, want %d", reserveResp.StatusCode, http.StatusUnauthorized)
	}
}

func mustRequest(t *testing.T, client *http.Client, method, url string, body []byte, headers map[string]string) *http.Response {
	t.Helper()

	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("create request %s %s: %v", method, url, err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("skipping e2e test: API is unavailable at %s: %v", url, err)
	}
	return resp
}
