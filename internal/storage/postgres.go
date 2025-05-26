package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
)

type PostgresStorage struct {
	db     *sql.DB
	logger *zap.Logger
}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func NewPostgresStorage(ctx context.Context, cfg Config, logger *zap.Logger) (*PostgresStorage, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	var db *sql.DB
	var err error

	// Configure retry policy
	retryPolicy := backoff.NewExponentialBackOff()
	retryPolicy.MaxElapsedTime = 2 * time.Minute
	retryPolicy.MaxInterval = 15 * time.Second

	logger.Info("Connecting to PostgreSQL...")

	err = backoff.RetryNotify(
		func() error {
			var innerErr error
			db, innerErr = sql.Open("postgres", connStr)
			if innerErr != nil {
				return fmt.Errorf("open postgres connection: %w", innerErr)
			}

			if innerErr = db.PingContext(ctx); innerErr != nil {
				return fmt.Errorf("ping postgres: %w", innerErr)
			}
			return nil
		},
		retryPolicy,
		func(err error, duration time.Duration) {
			logger.Warn("PostgreSQL connection failed, retrying...",
				zap.Error(err),
				zap.Duration("next_attempt_in", duration))
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL after retries: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.Info("Running database migrations...")
	if err := RunMigrations(ctx, db, "postgres"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL")
	return &PostgresStorage{
		db:     db,
		logger: logger,
	}, nil
}

func (s *PostgresStorage) SaveOrder(ctx context.Context, order Order) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Check if texture exists
	var textureExists bool
	err = tx.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM textures WHERE id = $1)",
		order.TextureID).Scan(&textureExists)
	if err != nil {
		return fmt.Errorf("check texture existence: %w", err)
	}
	if !textureExists {
		return fmt.Errorf("texture with id %s does not exist", order.TextureID)
	}

	// Insert order
	query := `INSERT INTO orders (
		user_id, width_cm, height_cm, texture_id, price, contact, 
		created_at, status
	) VALUES ($1, $2, $3, $4, $5, $6, NOW(), 'new')
	RETURNING id`

	var orderID int64
	err = tx.QueryRowContext(ctx, query,
		order.UserID,
		order.WidthCM,
		order.HeightCM,
		order.TextureID,
		order.Price,
		order.Contact,
	).Scan(&orderID)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	return tx.Commit()
}

func (s *PostgresStorage) GetOrders(ctx context.Context, limit, offset int) ([]Order, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
	
	query := `SELECT id, user_id, width_cm, height_cm, texture_id, price, 
		contact, created_at, status FROM orders 
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		err := rows.Scan(
			&o.ID,
			&o.UserID,
			&o.WidthCM,
			&o.HeightCM,
			&o.TextureID,
			&o.Price,
			&o.Contact,
			&o.CreatedAt,
			&o.Status,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

type Order struct {
	ID        int64
	UserID    int64
	WidthCM   int
	HeightCM  int
	TextureID string
	Price     float64
	Contact   string
	CreatedAt string
	Status    string
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
