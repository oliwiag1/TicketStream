package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type WSHandler struct {
	logger *log.Logger
	redis  *redis.Client
}

func NewWSHandler(logger *log.Logger, redis *redis.Client) *WSHandler {
	return &WSHandler{logger: logger, redis: redis}
}

func (h *WSHandler) Connect(c echo.Context) error {
	_ = h.redis
	return c.JSON(http.StatusNotImplemented, map[string]string{"message": "websocket stub"})
}
