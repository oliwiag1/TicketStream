package middleware

import (
	"net/http"
	"strings"

	"ticketstream/backend/internal/auth"

	"github.com/labstack/echo/v4"
)

const ClaimsContextKey = "auth.claims"

func RequireAuth(tokenValidator *auth.Validator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractBearerToken(c.Request().Header.Get("Authorization"))
			if token == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
			}

			claims, err := tokenValidator.Verify(c.Request().Context(), token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}

			c.Set(ClaimsContextKey, claims)
			return next(c)
		}
	}
}

func GetClaims(c echo.Context) (*auth.TokenClaims, bool) {
	value := c.Get(ClaimsContextKey)
	if value == nil {
		return nil, false
	}
	claims, ok := value.(*auth.TokenClaims)
	return claims, ok
}

func extractBearerToken(authorizationHeader string) string {
	if authorizationHeader == "" {
		return ""
	}

	parts := strings.SplitN(authorizationHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
