package middleware

import (
	"encoding/json"
	"log"
	"strings"

	"ticketstream/backend/internal/observability"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func RequestIDMiddleware() echo.MiddlewareFunc {
	return echomiddleware.RequestID()
}

func Logger(appLogger *log.Logger) echo.MiddlewareFunc {
	return echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogMethod:   true,
		LogRemoteIP: true,
		LogError:    true,
		LogLatency:  true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v echomiddleware.RequestLoggerValues) error {
			route := strings.TrimSpace(c.Path())
			if route == "" {
				route = v.URI
			}

			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			traceID := c.Response().Header().Get(TraceIDHeader)
			if traceID == "" {
				traceID = c.Request().Header.Get(TraceIDHeader)
			}

			observability.ObserveRequest(v.Method+" "+route, v.Status, v.Latency)

			entry := map[string]any{
				"type":       "request",
				"method":     v.Method,
				"uri":        v.URI,
				"route":      route,
				"status":     v.Status,
				"ip":         v.RemoteIP,
				"latency_ms": float64(v.Latency.Microseconds()) / 1000.0,
				"request_id": requestID,
				"trace_id":   traceID,
			}
			if v.Error != nil {
				entry["error"] = v.Error.Error()
			}

			payload, err := json.Marshal(entry)
			if err != nil {
				appLogger.Printf("request method=%s uri=%s status=%d ip=%s request_id=%s trace_id=%s", v.Method, v.URI, v.Status, v.RemoteIP, requestID, traceID)
				return nil
			}

			appLogger.Print(string(payload))
			return nil
		},
	})
}

func Recover() echo.MiddlewareFunc {
	return echomiddleware.Recover()
}
