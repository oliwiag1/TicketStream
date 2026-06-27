package middleware

import (
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const TraceIDHeader = "X-Trace-ID"

func TraceID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			traceID := strings.TrimSpace(c.Request().Header.Get(TraceIDHeader))
			if traceID == "" {
				traceID = strings.TrimSpace(c.Request().Header.Get(echo.HeaderXRequestID))
			}
			if traceID == "" {
				traceID = uuid.NewString()
			}

			c.Request().Header.Set(TraceIDHeader, traceID)
			c.Response().Header().Set(TraceIDHeader, traceID)
			return next(c)
		}
	}
}
