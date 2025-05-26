package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
)

func RunMigrations(ctx context.Context, db *sql.DB, dialect string) error {
	// Set the migrations base directory
	migrationsDir := filepath.Join("internal", "storage", "migrations")
	
	// Use OS filesystem but point to our migrations directory
	goose.SetBaseFS(os.DirFS(migrationsDir))
	
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

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
