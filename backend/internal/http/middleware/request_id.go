package middleware

import (
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func RequestID() echo.MiddlewareFunc {
	return echomiddleware.RequestID()
}
