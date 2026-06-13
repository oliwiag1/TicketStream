package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func ReserveRateLimit(maxAttempts int, _ int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			_ = maxAttempts
			return next(c)
		}
	}
}

func TooManyRequests(c echo.Context, message string) error {
	return c.JSON(http.StatusTooManyRequests, map[string]string{"error": message})
}
