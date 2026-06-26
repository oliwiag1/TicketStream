package routes

import (
	"log"

	"ticketstream/backend/internal/auth"
	"ticketstream/backend/internal/config"
	"ticketstream/backend/internal/http/handlers"
	httpmiddleware "ticketstream/backend/internal/http/middleware"
	"ticketstream/backend/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type Router struct {
	cfg       config.Config
	logger    *log.Logger
	pgPool    *pgxpool.Pool
	redis     *redis.Client
	validator *auth.Validator
}

func NewRouter(
	cfg config.Config,
	logger *log.Logger,
	pgPool *pgxpool.Pool,
	redis *redis.Client,
	validator *auth.Validator,
) *Router {
	return &Router{cfg: cfg, logger: logger, pgPool: pgPool, redis: redis, validator: validator}
}

func (r *Router) Register(e *echo.Echo) {
	realtimeService := services.NewRealtimeService(r.logger, r.pgPool, r.redis)
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(r.logger, r.cfg.KeycloakIssuerURL, r.cfg.KeycloakAudience)
	eventsHandler := handlers.NewEventsHandler(r.logger, r.pgPool, r.redis)
	reservationHandler := handlers.NewReservationHandler(r.cfg, r.logger, r.pgPool, r.redis, realtimeService)
	paymentHandler := handlers.NewPaymentHandler(r.logger, r.pgPool, r.redis, realtimeService)
	wsHandler := handlers.NewWSHandler(r.logger, realtimeService)
	requireAuth := httpmiddleware.RequireAuth(r.validator)

	e.GET("/health", healthHandler.Get)
	e.GET("/auth/provider", authHandler.ProviderInfo)
	e.GET("/auth/me", authHandler.Me, requireAuth)

	e.GET("/events", eventsHandler.List)
	e.POST("/events", eventsHandler.Create, requireAuth)
	e.GET("/events/:eventId/seats", eventsHandler.Seats)
	e.POST("/events/:eventId/reserve", reservationHandler.Reserve, requireAuth, httpmiddleware.ReserveRateLimit(r.cfg.RateLimitReserve, r.cfg.RateLimitWindowSecond))
	e.POST("/pay", paymentHandler.Pay, requireAuth)
	e.DELETE("/reservations/:reservationId", reservationHandler.Cancel, requireAuth)

	e.GET("/ws/events/:eventId", wsHandler.Connect)
}
