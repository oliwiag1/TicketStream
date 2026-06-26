package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ticketstream/backend/internal/auth"
	"ticketstream/backend/internal/config"
	httpmiddleware "ticketstream/backend/internal/http/middleware"
	"ticketstream/backend/internal/http/routes"
	"ticketstream/backend/pkg/broker"
	"ticketstream/backend/pkg/cache"
	"ticketstream/backend/pkg/db"
	"ticketstream/backend/pkg/logger"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.Load()
	appLogger := logger.New(cfg.AppEnv)

	ctx := context.Background()

	pgPool, err := db.NewPostgresPool(ctx, cfg.PostgresURL)
	if err != nil {
		appLogger.Fatalf("postgres connection failed: %v", err)
	}
	defer pgPool.Close()

	if err := db.RunMigrations(ctx, pgPool, cfg.MigrationsDir, appLogger); err != nil {
		appLogger.Fatalf("database migrations failed: %v", err)
	}

	redisClient := cache.NewRedisClient(cfg.RedisURL)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		appLogger.Fatalf("redis connection failed: %v", err)
	}
	defer redisClient.Close()

	rabbitConn, rabbitChannel, err := broker.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		appLogger.Fatalf("rabbitmq connection failed: %v", err)
	}
	defer rabbitConn.Close()
	defer rabbitChannel.Close()

	e := echo.New()
	e.HideBanner = true
	e.Use(httpmiddleware.RequestID())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderAuthorization, echo.HeaderContentType, echo.HeaderAccept, "Idempotency-Key"},
		AllowCredentials: true,
	}))
	e.Use(httpmiddleware.Logger(appLogger))
	e.Use(httpmiddleware.Recover())

	tokenValidator := auth.NewValidator(cfg.KeycloakIssuerURL, cfg.KeycloakAudience, cfg.KeycloakJWKSURL)

	router := routes.NewRouter(cfg, appLogger, pgPool, redisClient, rabbitChannel, tokenValidator)
	router.Register(e)

	go func() {
		addr := ":" + cfg.AppPort
		appLogger.Printf("api listening on %s", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			appLogger.Fatalf("api server failed: %v", err)
		}
	}()

	waitForShutdown(appLogger, e)
}

func waitForShutdown(appLogger *log.Logger, e *echo.Echo) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		appLogger.Printf("graceful shutdown failed: %v", err)
	}
}
