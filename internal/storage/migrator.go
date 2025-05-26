package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func RunMigrations(ctx context.Context, db *sql.DB, dialect string) error {
    if err := goose.SetDialect(dialect); err != nil {
        return fmt.Errorf("failed to set dialect: %w", err)
    }

    // Run migrations from the current directory
    if err := goose.UpContext(ctx, db, "."); err != nil {
        return fmt.Errorf("failed to run migrations: %w", err)
    }
    return nil
}

func RollbackMigration(ctx context.Context, db *sql.DB, dialect string) error {
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.DownContext(ctx, db, "."); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}
	return nil
}
