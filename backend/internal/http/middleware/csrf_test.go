package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCSRFProtectAllowsConfiguredOrigin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/pay", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := CSRFProtect([]string{"http://localhost:3000"})
	h := mw(func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })

	if err := h(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestCSRFProtectRejectsUnexpectedOrigin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/reservations/1", nil)
	req.Header.Set("Origin", "https://attacker.example")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := CSRFProtect([]string{"http://localhost:3000"})
	h := mw(func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })

	if err := h(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestCSRFProtectSkipsReadMethods(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := CSRFProtect([]string{"http://localhost:3000"})
	h := mw(func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })

	if err := h(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
