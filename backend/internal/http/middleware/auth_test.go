package middleware

import (
	"net/http/httptest"
	"testing"

	"ticketstream/backend/internal/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{name: "empty header", header: "", want: ""},
		{name: "missing space", header: "BearerToken", want: ""},
		{name: "wrong schema", header: "Basic abc", want: ""},
		{name: "bearer lowercase", header: "bearer token-123", want: "token-123"},
		{name: "trim token", header: "Bearer   token-xyz   ", want: "token-xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBearerToken(tt.header)
			if got != tt.want {
				t.Fatalf("extractBearerToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetClaims(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if claims, ok := GetClaims(c); ok || claims != nil {
		t.Fatalf("GetClaims() should return nil,false when context is empty")
	}

	expected := &auth.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "subject-1"},
		Subject:          "subject-1",
		Email:            "user@example.com",
	}
	c.Set(ClaimsContextKey, expected)

	claims, ok := GetClaims(c)
	if !ok || claims == nil {
		t.Fatalf("GetClaims() should return claims for populated context")
	}
	if claims.Subject != expected.Subject {
		t.Fatalf("GetClaims().Subject = %q, want %q", claims.Subject, expected.Subject)
	}
}
