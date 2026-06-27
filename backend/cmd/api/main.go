package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"ticketstream/backend/internal/auth"
	"ticketstream/backend/internal/config"
	httpmiddleware "ticketstream/backend/internal/http/middleware"
	"ticketstream/backend/internal/http/routes"
	"ticketstream/backend/internal/observability"
	"ticketstream/backend/internal/services"
	"ticketstream/backend/pkg/broker"
	"ticketstream/backend/pkg/cache"
	"ticketstream/backend/pkg/db"
	"ticketstream/backend/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
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
	e.Use(httpmiddleware.TraceID())
	e.Use(httpmiddleware.SecurityHeaders())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderAuthorization, echo.HeaderContentType, echo.HeaderAccept, "Idempotency-Key"},
		AllowCredentials: true,
	}))
	e.Use(echomiddleware.TimeoutWithConfig(echomiddleware.TimeoutConfig{
		Timeout: 30 * time.Second,
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/ws/")
		},
	}))
	e.Use(httpmiddleware.Logger(appLogger))
	e.Use(httpmiddleware.Recover())

	tokenValidator := auth.NewValidator(cfg.KeycloakIssuerURL, cfg.KeycloakAudience, cfg.KeycloakJWKSURL)
	realtimeService := services.NewRealtimeService(appLogger, pgPool, redisClient)

	router := routes.NewRouter(cfg, appLogger, pgPool, redisClient, tokenValidator, realtimeService)
	router.Register(e)

	outboxService := services.NewOutboxService(pgPool, rabbitChannel, appLogger)
	reservationExpiryService := services.NewReservationExpiryService(appLogger, pgPool, redisClient, realtimeService)
	backgroundCtx, backgroundCancel := context.WithCancel(context.Background())
	defer backgroundCancel()
	go startOutboxPublisher(backgroundCtx, appLogger, outboxService)
	go startReservationExpirySweeper(backgroundCtx, appLogger, reservationExpiryService, cfg.ExpiredSweepSeconds)
	go startQueueLagSampler(backgroundCtx, appLogger, pgPool)

	go func() {
		addr := ":" + cfg.AppPort
		appLogger.Printf("api listening on %s", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			appLogger.Fatalf("api server failed: %v", err)
		}
	}()

	waitForShutdown(appLogger, e)
}

func startOutboxPublisher(ctx context.Context, appLogger *log.Logger, outboxService *services.OutboxService) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		if err := outboxService.PublishPending(ctx); err != nil {
			appLogger.Printf("outbox publish failed: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func startReservationExpirySweeper(ctx context.Context, appLogger *log.Logger, sweepService *services.ReservationExpiryService, everySeconds int) {
	if everySeconds <= 0 {
		everySeconds = 5
	}

	ticker := time.NewTicker(time.Duration(everySeconds) * time.Second)
	defer ticker.Stop()

	for {
		released, err := sweepService.ReleaseExpired(ctx, 100)
		if err != nil {
			appLogger.Printf("reservation expiry sweep failed: %v", err)
		} else if released > 0 {
			appLogger.Printf("reservation expiry sweep released=%d", released)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func startQueueLagSampler(ctx context.Context, appLogger *log.Logger, pgPool *pgxpool.Pool) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		var lag int64
		err := pgPool.QueryRow(ctx, `SELECT COUNT(*) FROM outbox_events WHERE published_at IS NULL`).Scan(&lag)
		if err != nil {
			appLogger.Printf("queue lag sample failed: %v", err)
		} else {
			observability.SetQueueLag(lag)
			appLogger.Print("{\"type\":\"queue_lag\",\"value\":" + strconv.FormatInt(lag, 10) + "}")
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
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
