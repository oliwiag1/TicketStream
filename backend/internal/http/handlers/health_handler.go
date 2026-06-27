package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	pgPool *pgxpool.Pool
	redis  *redis.Client
}

func NewHealthHandler(pgPool *pgxpool.Pool, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{pgPool: pgPool, redis: redisClient}
}

func (h *HealthHandler) Get(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Live(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "live"})
}

func (h *HealthHandler) Ready(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	if h.pgPool != nil {
		if err := h.pgPool.Ping(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "dependency": "postgres"})
		}
	}

	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "dependency": "redis"})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}
