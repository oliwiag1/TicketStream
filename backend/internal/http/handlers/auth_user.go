package handlers

import (
	"context"
	"errors"
	"strings"

	httpmiddleware "ticketstream/backend/internal/http/middleware"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func resolveAuthenticatedUser(ctx context.Context, c echo.Context, pgPool *pgxpool.Pool) (subject string, userID string, email string, err error) {
	claims, ok := httpmiddleware.GetClaims(c)
	if !ok || claims == nil || claims.Subject == "" {
		return "", "", "", errors.New("missing claims subject")
	}

	subject = claims.Subject
	userID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(subject)).String()
	email = strings.TrimSpace(claims.Email)
	if email == "" {
		email = userID + "@keycloak.local"
	}

	_, execErr := pgPool.Exec(
		ctx,
		`INSERT INTO users (id, email, password_hash)
		 VALUES ($1, $2, 'keycloak')
		 ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email`,
		userID,
		email,
	)
	if execErr != nil {
		return "", "", "", execErr
	}

	return subject, userID, email, nil
}
