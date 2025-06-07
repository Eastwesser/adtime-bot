package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

func RunMigrations(ctx context.Context, db *sql.DB, logger *zap.Logger) error {
	const operation = "storage.RunMigrations"
	
	logger.Info("Running database migrations...")

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("%s: failed to set dialect: %w", operation, err)
	}

	// Run migrations from embedded files or directory
	if err := goose.UpContext(ctx, db, "internal/storage/migrations"); err != nil {
		return fmt.Errorf("%s: failed to run migrations: %w", operation, err)
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

func RollbackMigration(ctx context.Context, db *sql.DB, logger *zap.Logger) error {
	const operation = "storage.RollbackMigration"
	
	logger.Info("Rolling back last migration...")

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("%s: failed to set dialect: %w", operation, err)
	}

	if err := goose.DownContext(ctx, db, "internal/storage/migrations"); err != nil {
		return fmt.Errorf("%s: failed to rollback migration: %w", operation, err)
	}

	logger.Info("Migration rollback completed")
	return nil
}

func Status(ctx context.Context, db *sql.DB, logger *zap.Logger) error {
	const operation = "storage.Status"
	
	logger.Info("Checking migration status...")

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("%s: failed to set dialect: %w", operation, err)
	}

	if err := goose.StatusContext(ctx, db, "internal/storage/migrations"); err != nil {
		return fmt.Errorf("%s: failed to check migration status: %w", operation, err)
	}

	return nil
}
