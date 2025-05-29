package storage

import (
	"adtime-bot/internal/config"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type PostgresStorage struct {
	db     *sqlx.DB
	logger *zap.Logger
}

type Texture struct {
	ID          string  `db:"id"`
	Name        string  `db:"name"`
	PricePerDM2 float64 `db:"price_per_dm2"`
	ImageURL    string  `db:"image_url"`
	InStock     bool    `db:"in_stock"`
}

type Order struct {
    ID           int64     `db:"id"`
    UserID       int64     `db:"user_id"`
    WidthCM      int       `db:"width_cm"`
    HeightCM     int       `db:"height_cm"`
    TextureID    string    `db:"texture_id"`
    TextureName  string    `db:"-"`
    Price        float64   `db:"price"`
    LeatherCost  float64   `db:"leather_cost"`
    ProcessCost  float64   `db:"process_cost"`
    TotalCost    float64   `db:"total_cost"`
    Commission   float64   `db:"commission"`
    Tax          float64   `db:"tax"`
    NetRevenue   float64   `db:"net_revenue"`
    Profit       float64   `db:"profit"`
    Contact      string    `db:"contact"`
    Status       string    `db:"status"`
    CreatedAt    time.Time `db:"created_at"`
}

type PriceFormula struct {
    ID          string
    ServiceType string
    Formula     string // "width*height*price*coefficient"
    Parameters  map[string]float64
}

func NewPostgresStorage(ctx context.Context, cfg config.Config, logger *zap.Logger) (*PostgresStorage, error) {
	const operation = "storage.NewPostgresStorage"

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	var db *sqlx.DB
	var err error

	retryPolicy := backoff.NewExponentialBackOff()
	retryPolicy.MaxElapsedTime = 2 * time.Minute
	retryPolicy.MaxInterval = 15 * time.Second

	logger.Info("Connecting to PostgreSQL...")

	err = backoff.RetryNotify(
		func() error {
			db, err = sqlx.ConnectContext(ctx, "postgres", connStr)
			if err != nil {
				return fmt.Errorf("connect: %w", err)
			}

			if err = db.PingContext(ctx); err != nil {
				return fmt.Errorf("ping: %w", err)
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
		return nil, fmt.Errorf("%s: failed to connect after retries: %w", operation, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	logger.Info("Successfully connected to PostgreSQL")
	return &PostgresStorage{db: db, logger: logger}, nil
}

func (s *PostgresStorage) GetTextureByID(ctx context.Context, textureID string) (*Texture, error) {
	const query = `
		SELECT id::text, name, price_per_dm2, image_url, in_stock 
		FROM textures 
		WHERE id = $1
	`

	var texture Texture
	err := s.db.GetContext(ctx, &texture, query, textureID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("texture not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get texture: %w", err)
	}

	return &texture, nil
}

func (s *PostgresStorage) GetAvailableTextures(ctx context.Context) ([]Texture, error) {
	const query = `SELECT id::text, name, price_per_dm2, image_url FROM textures WHERE in_stock = TRUE`
	
	var textures []Texture
	err := s.db.SelectContext(ctx, &textures, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get textures: %w", err)
	}
	
	return textures, nil
}

func (s *PostgresStorage) SaveOrder(ctx context.Context, order Order) (int64, error) {
    const query = `
        INSERT INTO orders (
            user_id, width_cm, height_cm, texture_id, price,
            leather_cost, process_cost, total_cost, commission,
            tax, net_revenue, profit, contact, status, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
        RETURNING id
    `

    var orderID int64
    err := s.db.QueryRowContext(ctx, query,
        order.UserID,
        order.WidthCM,
        order.HeightCM,
        order.TextureID,
        order.Price,
        order.LeatherCost,
        order.ProcessCost,
        order.TotalCost,
        order.Commission,
        order.Tax,
        order.NetRevenue,
        order.Profit,
        order.Contact,
        order.Status,
        order.CreatedAt,
    ).Scan(&orderID)

    if err != nil {
        return 0, fmt.Errorf("failed to save order: %w", err)
    }

    return orderID, nil
}

func (s *PostgresStorage) ExportOrderToExcel(ctx context.Context, order Order) error {
    f := excelize.NewFile()
    defer f.Close()

    index, err := f.NewSheet("Orders")
    if err != nil {
        return fmt.Errorf("failed to create sheet: %w", err)
    }

    // Set headers
    headers := []string{
        "ID", "User ID", "Width (cm)", "Height (cm)",
        "Texture ID", "Texture Name", "Price",
        "Leather Cost", "Process Cost", "Total Cost",
        "Commission", "Tax", "Net Revenue", "Profit",
        "Contact", "Status", "Created At",
    }
    for i, header := range headers {
        cell, _ := excelize.CoordinatesToCellName(i+1, 1)
        f.SetCellValue("Orders", cell, header)
    }

    // Set data
    data := []interface{}{
        order.ID,
        order.UserID,
        order.WidthCM,
        order.HeightCM,
        order.TextureID,
        order.TextureName,
        order.Price,
        order.LeatherCost,
        order.ProcessCost,
        order.TotalCost,
        order.Commission,
        order.Tax,
        order.NetRevenue,
        order.Profit,
        order.Contact,
        order.Status,
        order.CreatedAt.Format("2006-01-02 15:04:05"),
    }
    for i, value := range data {
        cell, _ := excelize.CoordinatesToCellName(i+1, 2)
        f.SetCellValue("Orders", cell, value)
    }

    // Calculate and add area
    area := float64(order.WidthCM*order.HeightCM) / 100
    f.SetCellValue("Orders", "R1", "Area (dm²)")
    f.SetCellValue("Orders", "R2", area)

    f.SetActiveSheet(index)

    // Ensure orders directory exists
    if err := os.MkdirAll("orders", 0755); err != nil {
        return fmt.Errorf("failed to create orders directory: %w", err)
    }

    // Generate filename
    filename := filepath.Join("orders",
        fmt.Sprintf("order_%d_%s.xlsx",
            order.ID,
            order.CreatedAt.Format("20060102_150405")))

    if err := f.SaveAs(filename); err != nil {
        return fmt.Errorf("failed to save Excel file: %w", err)
    }

    return nil
}

func (s *PostgresStorage) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}
