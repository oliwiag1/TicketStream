package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pgPool *pgxpool.Pool, migrationsDir string, appLogger *log.Logger) error {
	if strings.TrimSpace(migrationsDir) == "" {
		migrationsDir = "migrations"
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	migrationFiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	if _, err := pgPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, fileName := range migrationFiles {
		alreadyApplied, err := isMigrationApplied(ctx, pgPool, fileName)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", fileName, err)
		}
		if alreadyApplied {
			continue
		}

		fullPath := filepath.Join(migrationsDir, fileName)
		sqlBytes, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", fileName, err)
		}

		sqlText := strings.TrimSpace(string(sqlBytes))
		if sqlText == "" {
			if _, err := pgPool.Exec(ctx, `INSERT INTO schema_migrations (name) VALUES ($1)`, fileName); err != nil {
				return fmt.Errorf("mark empty migration %s: %w", fileName, err)
			}
			continue
		}

		tx, err := pgPool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration tx %s: %w", fileName, err)
		}

		if _, err := tx.Exec(ctx, sqlText); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("execute migration %s: %w", fileName, err)
		}

		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (name) VALUES ($1)`, fileName); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", fileName, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", fileName, err)
		}

		if appLogger != nil {
			appLogger.Printf("migration applied: %s", fileName)
		}
	}

	return nil
}

func isMigrationApplied(ctx context.Context, pgPool *pgxpool.Pool, name string) (bool, error) {
	var exists bool
	err := pgPool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)`, name).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return exists, nil
}
