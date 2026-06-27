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
	realtime  *services.RealtimeService
}

func NewRouter(
	cfg config.Config,
	logger *log.Logger,
	pgPool *pgxpool.Pool,
	redis *redis.Client,
	validator *auth.Validator,
	realtime *services.RealtimeService,
) *Router {
	return &Router{cfg: cfg, logger: logger, pgPool: pgPool, redis: redis, validator: validator, realtime: realtime}
}

func (r *Router) Register(e *echo.Echo) {
	realtimeService := r.realtime
	if realtimeService == nil {
		realtimeService = services.NewRealtimeService(r.logger, r.pgPool, r.redis)
	}
	healthHandler := handlers.NewHealthHandler(r.pgPool, r.redis)
	authHandler := handlers.NewAuthHandler(r.logger, r.cfg.KeycloakIssuerURL, r.cfg.KeycloakAudience)
	eventsHandler := handlers.NewEventsHandler(r.logger, r.pgPool, r.redis)
	reservationHandler := handlers.NewReservationHandler(r.cfg, r.logger, r.pgPool, r.redis, realtimeService)
	paymentHandler := handlers.NewPaymentHandler(r.logger, r.pgPool, r.redis, realtimeService)
	ticketsHandler := handlers.NewTicketsHandler(r.logger, r.pgPool)
	wsHandler := handlers.NewWSHandler(r.logger, realtimeService)
	metricsHandler := handlers.NewMetricsHandler()
	requireAuth := httpmiddleware.RequireAuth(r.validator)
	csrfProtect := httpmiddleware.CSRFProtect(r.cfg.CORSAllowOrigins)
	authRateLimit := httpmiddleware.AuthRateLimit(r.redis, r.cfg.RateLimitAuth, r.cfg.RateLimitWindowSecond)
	reserveRateLimit := httpmiddleware.ReserveRateLimit(r.redis, r.cfg.RateLimitReserve, r.cfg.RateLimitWindowSecond)

	e.GET("/health", healthHandler.Get)
	e.GET("/health/live", healthHandler.Live)
	e.GET("/health/ready", healthHandler.Ready)
	e.GET("/metrics", metricsHandler.JSON)
	e.GET("/ops/dashboard", metricsHandler.Dashboard)
	e.GET("/auth/provider", authHandler.ProviderInfo, authRateLimit)
	e.GET("/auth/me", authHandler.Me, authRateLimit, requireAuth)

	e.GET("/events", eventsHandler.List)
	e.POST("/events", eventsHandler.Create, requireAuth, csrfProtect)
	e.GET("/events/:eventId/seats", eventsHandler.Seats)
	e.POST("/events/:eventId/reserve", reservationHandler.Reserve, requireAuth, csrfProtect, reserveRateLimit)
	e.POST("/pay", paymentHandler.Pay, requireAuth, csrfProtect)
	e.DELETE("/reservations/:reservationId", reservationHandler.Cancel, requireAuth, csrfProtect)
	e.GET("/tickets/me", ticketsHandler.ListMine, requireAuth)

	e.GET("/ws/events/:eventId", wsHandler.Connect)
}
