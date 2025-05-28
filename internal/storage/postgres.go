package storage

import (
	"adtime-bot/internal/config"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"os"
	"path/filepath"

	"github.com/cenkalti/backoff/v4"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

// Texture represents a product texture
type Texture struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    PricePerDM2 float64 `json:"price_per_dm2"`
    ImageURL    string  `json:"image_url"`
    InStock     bool    `json:"in_stock"`
}

// Order represents a customer order
type Order struct {
    ID          int64     `json:"id"`
    UserID      int64     `json:"user_id"`
    WidthCM     int       `json:"width_cm"`
    HeightCM    int       `json:"height_cm"`
    TextureID   string    `json:"texture_id"`
    TextureName string    `json:"texture_name"`
    PricePerDM2 float64   `json:"price_per_dm2"`
    TotalPrice  float64   `json:"total_price"`
    Contact     string    `json:"contact"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
}

type PostgresStorage struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPostgresStorage(ctx context.Context, cfg config.Config, logger *zap.Logger) (*PostgresStorage, error) {
    
    connStr := fmt.Sprintf(
        // "host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", // Changed sslmode=require to sslmode=disable
        cfg.Database.Host, 
        cfg.Database.Port, 
        cfg.Database.User, 
        cfg.Database.Password, 
        cfg.Database.Name,
    )

	var db *sql.DB
	var err error

	retryPolicy := backoff.NewExponentialBackOff()
	retryPolicy.MaxElapsedTime = 2 * time.Minute
	retryPolicy.MaxInterval = 15 * time.Second

	logger.Info("Connecting to PostgreSQL...")

	err = backoff.RetryNotify(
		func() error {
			db, err = sql.Open("postgres", connStr)
			if err != nil {
				return fmt.Errorf("open connection: %w", err)
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
		return nil, fmt.Errorf("failed to connect after retries: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	logger.Info("Running database migrations...")
	if err := RunMigrations(ctx, db, "postgres"); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL")
	return &PostgresStorage{db: db, logger: logger}, nil
}

// func (s *PostgresStorage) GetAvailableTextures(ctx context.Context) ([]Texture, error) {
//     const query = `SELECT id, name, price_per_dm2, image_url, in_stock FROM textures WHERE in_stock = TRUE`

//     rows, err := s.db.QueryContext(ctx, query)
//     if err != nil {
//         return nil, fmt.Errorf("query textures: %w", err)
//     }
//     defer func() {
//         if err := rows.Close(); err != nil {
//             s.logger.Error("Failed to close rows", zap.Error(err))
//         }
//     }()

//     var textures []Texture
//     for rows.Next() {
//         var t Texture
//         if err := rows.Scan(
//             &t.ID,
//             &t.Name,
//             &t.PricePerDM2,
//             &t.ImageURL,
//             &t.InStock,
//         ); err != nil {
//             return nil, fmt.Errorf("scan texture: %w", err)
//         }
//         textures = append(textures, t)
//     }
//     if err := rows.Err(); err != nil {
//         return nil, fmt.Errorf("rows error: %w", err)
//     }

//     return textures, nil
// }

func (s *PostgresStorage) SaveOrder(ctx context.Context, order Order) (int64, error) {
    const query = `
        INSERT INTO orders (
            user_id, width_cm, height_cm, 
            texture_id, price, 
            contact, status, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `

    var orderID int64
    err := s.db.QueryRowContext(ctx, query,
        order.UserID,
        order.WidthCM,
        order.HeightCM,
        order.TextureID,
        order.TotalPrice,
        order.Contact,
        order.Status,
        order.CreatedAt,
    ).Scan(&orderID)

    if err != nil {
        return 0, fmt.Errorf("insert order: %w", err)
    }

    return orderID, nil
}

// func (s *PostgresStorage) GetOrders(ctx context.Context, limit, offset int) ([]Order, error) {
//     const query = `
//         SELECT 
//             o.id, o.user_id, o.width_cm, o.height_cm, 
//             o.texture_id, t.name as texture_name, t.price_per_dm2,
//             o.price as total_price, 
//             o.contact, o.status, o.created_at
//         FROM orders o
//         JOIN textures t ON o.texture_id = t.id
//         ORDER BY o.created_at DESC
//         LIMIT $1 OFFSET $2
//     `

//     rows, err := s.db.QueryContext(ctx, query, limit, offset)
//     if err != nil {
//         return nil, fmt.Errorf("query orders: %w", err)
//     }
//     defer rows.Close()

//     var orders []Order
//     for rows.Next() {
//         var o Order
//         err := rows.Scan(
//             &o.ID,
//             &o.UserID,
//             &o.WidthCM,
//             &o.HeightCM,
//             &o.TextureID,
//             &o.TextureName,
//             &o.PricePerDM2,
//             &o.TotalPrice,
//             &o.Contact,
//             &o.Status,
//             &o.CreatedAt,
//         )
//         if err != nil {
//             return nil, fmt.Errorf("scan order: %w", err)
//         }
//         orders = append(orders, o)
//     }
//     if err := rows.Err(); err != nil {
//         return nil, fmt.Errorf("rows error: %w", err)
//     }

//     return orders, nil
// }

func (s *PostgresStorage) GetTextureByID(ctx context.Context, textureID string) (*Texture, error) {
    const query = `
        SELECT id::text, name, price_per_dm2, image_url, in_stock 
        FROM textures 
        WHERE id = $1
    `

    var texture Texture
    err := s.db.QueryRowContext(ctx, query, textureID).Scan(
        &texture.ID,
        &texture.Name,
        &texture.PricePerDM2,
        &texture.ImageURL,
        &texture.InStock,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("texture not found: %w", err)
        }
        return nil, fmt.Errorf("query texture: %w", err)
    }

    return &texture, nil
}

func (s *PostgresStorage) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *PostgresStorage) ExportOrderToExcel(ctx context.Context, order Order) error {
	f := excelize.NewFile()
	index, err := f.NewSheet("Orders")
	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}

	headers := []string{
		"ID", "User ID", "Width (cm)", "Height (cm)",
		"Texture ID", "Texture Name", "Price per dm²",
		"Total Price", "Contact", "Status", "Created At",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Orders", cell, header)
	}

	data := []interface{}{
		order.ID,
		order.UserID,
		order.WidthCM,
		order.HeightCM,
		order.TextureID,
		order.TextureName,
		order.PricePerDM2,
		order.TotalPrice,
		order.Contact,
		order.Status,
		order.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	for i, value := range data {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue("Orders", cell, value)
	}

	area := float64(order.WidthCM*order.HeightCM) / 100
	f.SetCellValue("Orders", "K1", "Area (dm²)")
	f.SetCellValue("Orders", "K2", area)

	f.SetActiveSheet(index)

	if err := os.MkdirAll("orders", 0755); err != nil {
		return fmt.Errorf("failed to create orders directory: %w", err)
	}

	filename := filepath.Join("orders",
		fmt.Sprintf("order_%d_%s.xlsx",
			order.ID,
			order.CreatedAt.Format("20060102_150405")))

	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("failed to save Excel file: %w", err)
	}

	return nil
}
