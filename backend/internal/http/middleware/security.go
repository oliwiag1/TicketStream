package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			headers := c.Response().Header()
			headers.Set(echo.HeaderXContentTypeOptions, "nosniff")
			headers.Set(echo.HeaderXFrameOptions, "DENY")
			headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			headers.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			headers.Set("Cache-Control", "no-store")
			headers.Set("Pragma", "no-cache")

			if c.IsTLS() {
				headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			if c.Request().Method == http.MethodGet {
				headers.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
			}

			return next(c)
		}
	}
}
