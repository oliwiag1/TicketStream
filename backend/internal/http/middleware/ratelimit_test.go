package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ticketstream/backend/internal/auth"

	"github.com/labstack/echo/v4"
)

func TestRateLimitKeyUsesSubjectWhenAvailable(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/events/abc/reserve", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/events/:eventId/reserve")
	c.Set(ClaimsContextKey, &auth.TokenClaims{Subject: "user-123"})

	key := rateLimitKey("reserve", c)
	if !strings.Contains(key, "user-123") {
		t.Fatalf("rateLimitKey should contain claims subject, got %s", key)
	}
}

func TestRateLimitKeyFallsBackToIP(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.RemoteAddr = "192.168.1.20:1234"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/auth/me")

	key := rateLimitKey("auth", c)
	if !strings.Contains(key, "192.168.1.20") {
		t.Fatalf("rateLimitKey should contain remote ip, got %s", key)
	}
}
