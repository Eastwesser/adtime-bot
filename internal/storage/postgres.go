package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
	"github.com/xuri/excelize/v2"
    "os"
    "path/filepath"
)

type PostgresStorage struct {
	db     *sql.DB
	logger *zap.Logger
}

func (s *PostgresStorage) GetAvailableTextures(ctx context.Context) ([]Texture, error) {
    const query = `SELECT id, name, price_per_dm2, image_url FROM textures WHERE in_stock = TRUE`
    rows, err := s.db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer rows.Close()

    var textures []Texture
    for rows.Next() {
        var t Texture
        if err := rows.Scan(&t.ID, &t.Name, &t.PricePerDM2, &t.ImageURL); err != nil {
            return nil, fmt.Errorf("scan failed: %w", err)
        }
        textures = append(textures, t)
    }
    return textures, nil
}

func (s *PostgresStorage) ExportOrderToExcel(ctx context.Context, order Order) error {
    // Implement Excel export logic here
    // This could use a library like excelize or tealeg/xlsx
    return nil
}

type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// NewPostgresStorage creates a new PostgreSQL storage instance with retry logic
func NewPostgresStorage(ctx context.Context, cfg Config, logger *zap.Logger) (*PostgresStorage, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

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
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	logger.Info("Running database migrations...")
	if err := RunMigrations(ctx, db, "postgres"); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL")
	return &PostgresStorage{db: db, logger: logger}, nil
}

// SaveOrder saves a new order with transaction and validation
func (s *PostgresStorage) SaveOrder(ctx context.Context, order Order) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    defer func() {
        if err != nil {
            if rbErr := tx.Rollback(); rbErr != nil {
                s.logger.Error("Failed to rollback transaction", zap.Error(rbErr))
            }
        }
    }()

    // Insert order with all fields
    const query = `INSERT INTO orders (
        user_id, width_cm, height_cm, texture_id, texture_name, 
        price_per_dm2, total_price, contact, status, created_at
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    RETURNING id`

    var orderID int64
    err = tx.QueryRowContext(ctx, query,
        order.UserID,
        order.WidthCM,
        order.HeightCM,
        order.TextureID,
        order.TextureName,
        order.PricePerDM2,
        order.TotalPrice,
        order.Contact,
        order.Status,
        order.CreatedAt,
    ).Scan(&orderID)
    if err != nil {
        return fmt.Errorf("insert order failed: %w", err)
    }

    if err = tx.Commit(); err != nil {
        return fmt.Errorf("commit failed: %w", err)
    }

    return nil
}

// GetOrders retrieves a paginated list of orders
func (s *PostgresStorage) GetOrders(ctx context.Context, limit, offset int) ([]Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const query = `SELECT id, user_id, width_cm, height_cm, texture_id, price, 
		contact, created_at, status FROM orders 
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(
			&o.ID,
			&o.UserID,
			&o.WidthCM,
			&o.HeightCM,
			&o.TextureID,
			&o.Price,
			&o.Contact,
			&o.CreatedAt,
			&o.Status,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// GetTextureByID retrieves a single texture by ID
func (s *PostgresStorage) GetTextureByID(ctx context.Context, textureID string) (*Texture, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const query = `SELECT id, name, price_per_dm2, image_url, in_stock 
		FROM textures WHERE id = $1`

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
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &texture, nil
}

// Close closes the database connection
func (s *PostgresStorage) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Texture represents a product texture
type Texture struct {
    ID          string
    Name        string
    PricePerDM2 float64
    ImageURL    string
    InStock     bool
}

// Order represents a customer order
type Order struct {
    ID          int64
    UserID      int64
    WidthCM     int
    HeightCM    int
    TextureID   string
    TextureName string
    PricePerDM2 float64
    TotalPrice  float64
    Contact     string
    Status      string
    CreatedAt   time.Time
}

func (s *PostgresStorage) ExportOrderToExcel(ctx context.Context, order storage.Order) error {
    // Create a new Excel file
    f := excelize.NewFile()
    
    // Create a new sheet
    index, err := f.NewSheet("Orders")
    if err != nil {
        return fmt.Errorf("failed to create sheet: %w", err)
    }
    
    // Set headers
    headers := []string{
        "ID", "User ID", "Width (cm)", "Height (cm)", 
        "Texture ID", "Texture Name", "Price per dm²", 
        "Total Price", "Contact", "Status", "Created At",
    }
    
    for i, header := range headers {
        cell, _ := excelize.CoordinatesToCellName(i+1, 1)
        f.SetCellValue("Orders", cell, header)
    }
    
    // Set order data
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
    
    // Calculate area in dm²
    area := float64(order.WidthCM*order.HeightCM) / 100
    f.SetCellValue("Orders", "K1", "Area (dm²)")
    f.SetCellValue("Orders", "K2", area)
    
    // Set active sheet and save
    f.SetActiveSheet(index)
    
    // Create orders directory if it doesn't exist
    if err := os.MkdirAll("orders", 0755); err != nil {
        return fmt.Errorf("failed to create orders directory: %w", err)
    }
    
    // Generate filename with timestamp
    filename := filepath.Join("orders", 
        fmt.Sprintf("order_%d_%s.xlsx", 
            order.ID, 
            order.CreatedAt.Format("20060102_150405")))
    
    if err := f.SaveAs(filename); err != nil {
        return fmt.Errorf("failed to save Excel file: %w", err)
    }
    
    return nil
}

