package middleware

import (
	"log"

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
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v echomiddleware.RequestLoggerValues) error {
			if v.Error != nil {
				appLogger.Printf("request method=%s uri=%s status=%d ip=%s err=%v", v.Method, v.URI, v.Status, v.RemoteIP, v.Error)
				return nil
			}
			appLogger.Printf("request method=%s uri=%s status=%d ip=%s", v.Method, v.URI, v.Status, v.RemoteIP)
			return nil
		},
	})
}

func Recover() echo.MiddlewareFunc {
	return echomiddleware.Recover()
}
