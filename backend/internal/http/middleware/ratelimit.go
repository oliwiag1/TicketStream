package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

func ReserveRateLimit(redisClient *redis.Client, maxAttempts, windowSeconds int) echo.MiddlewareFunc {
	return RateLimit(redisClient, "reserve", maxAttempts, windowSeconds)
}

func AuthRateLimit(redisClient *redis.Client, maxAttempts, windowSeconds int) echo.MiddlewareFunc {
	return RateLimit(redisClient, "auth", maxAttempts, windowSeconds)
}

func RateLimit(redisClient *redis.Client, scope string, maxAttempts, windowSeconds int) echo.MiddlewareFunc {
	if maxAttempts <= 0 || windowSeconds <= 0 {
		return passthroughMiddleware
	}

	window := time.Duration(windowSeconds) * time.Second

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if redisClient == nil {
				return next(c)
			}

			key := rateLimitKey(scope, c)
			ctx := c.Request().Context()

			count, err := redisClient.Incr(ctx, key).Result()
			if err != nil {
				return next(c)
			}
			if count == 1 {
				_ = redisClient.Expire(ctx, key, window).Err()
			}

			if int(count) > maxAttempts {
				ttl, ttlErr := redisClient.TTL(ctx, key).Result()
				if ttlErr == nil && ttl > 0 {
					c.Response().Header().Set(echo.HeaderRetryAfter, strconv.Itoa(int(ttl.Seconds())))
				}
				return TooManyRequests(c, "rate_limit_exceeded")
			}

			return next(c)
		}
	}
}

var passthroughMiddleware = func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

func rateLimitKey(scope string, c echo.Context) string {
	identity := c.RealIP()
	if claims, ok := GetClaims(c); ok && claims != nil && strings.TrimSpace(claims.Subject) != "" {
		identity = claims.Subject
	}

	path := strings.TrimSpace(c.Path())
	if path == "" {
		path = c.Request().URL.Path
	}
	path = strings.ReplaceAll(path, "/", ":")

	return fmt.Sprintf("rate_limit:%s:%s:%s:%s", scope, c.Request().Method, path, identity)
}

func TooManyRequests(c echo.Context, message string) error {
	return c.JSON(http.StatusTooManyRequests, map[string]string{"error": message})
}
