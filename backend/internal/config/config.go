package config

import (
	"os"
	"strconv"
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
