package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppPort               string
	AppEnv                string
	MigrationsDir         string
	PostgresURL           string
	RedisURL              string
	RabbitMQURL           string
	KeycloakIssuerURL     string
	KeycloakAudience      string
	KeycloakJWKSURL       string
	CORSAllowOrigins      []string
	SeatLockTTLSeconds    int
	RateLimitReserve      int
	RateLimitWindowSecond int
}

func Load() Config {
	return Config{
		AppPort:               getenv("APP_PORT", "8080"),
		AppEnv:                getenv("APP_ENV", "dev"),
		MigrationsDir:         getenv("MIGRATIONS_DIR", "migrations"),
		PostgresURL:           getenv("POSTGRES_URL", "postgres://ticketstream:ticketstream@localhost:5432/ticketstream?sslmode=disable"),
		RedisURL:              getenv("REDIS_URL", "redis://localhost:6379/0"),
		RabbitMQURL:           getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		KeycloakIssuerURL:     getenv("KEYCLOAK_ISSUER_URL", "http://localhost:18081/realms/ticketstream"),
		KeycloakAudience:      getenv("KEYCLOAK_AUDIENCE", "ticketstream-frontend"),
		KeycloakJWKSURL:       getenv("KEYCLOAK_JWKS_URL", "http://host.docker.internal:18081/realms/ticketstream/protocol/openid-connect/certs"),
		CORSAllowOrigins:      getenvCSV("CORS_ALLOW_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173,http://localhost:3000,http://127.0.0.1:3000"),
		SeatLockTTLSeconds:    getenvInt("SEAT_LOCK_TTL_SECONDS", 600),
		RateLimitReserve:      getenvInt("RATE_LIMIT_RESERVE", 5),
		RateLimitWindowSecond: getenvInt("RATE_LIMIT_WINDOW_SECONDS", 10),
	}
}

func getenv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) int {
	value := getenv(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvCSV(key, fallback string) []string {
	raw := getenv(key, fallback)
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}
