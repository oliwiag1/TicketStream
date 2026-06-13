package handlers

import (
	"log"
	"net/http"

	httpmiddleware "ticketstream/backend/internal/http/middleware"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	logger *log.Logger
	issuer string
	aud    string
}

func NewAuthHandler(logger *log.Logger, issuer, audience string) *AuthHandler {
	return &AuthHandler{logger: logger, issuer: issuer, aud: audience}
}

func (h *AuthHandler) ProviderInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"provider": "keycloak",
		"issuer":   h.issuer,
		"audience": h.aud,
	})
}

func (h *AuthHandler) Me(c echo.Context) error {
	claims, ok := httpmiddleware.GetClaims(c)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing claims"})
	}

	roles := claims.RealmAccess.Roles
	return c.JSON(http.StatusOK, map[string]any{
		"sub":                claims.Subject,
		"preferred_username": claims.PreferredUsername,
		"email":              claims.Email,
		"roles":              roles,
		"scope":              claims.Scope,
	})
}
