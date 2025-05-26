package api

// API CLIENT

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	logger     *zap.Logger
}

type Texture struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	PricePerDM2 float64 `json:"price_per_dm2"`
	ImageURL    string  `json:"image_url"`
	InStock     bool    `json:"in_stock"`
}

type OrderRequest struct {
	UserID    int64   `json:"user_id"`
	WidthCM   int     `json:"width_cm"`
	HeightCM  int     `json:"height_cm"`
	TextureID string  `json:"texture_id"`
	Price     float64 `json:"price"`
	Contact   string  `json:"contact"`
}

func NewClient(baseURL, token string, logger *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) GetTextures(ctx context.Context) ([]Texture, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/api/textures", c.baseURL),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var textures []Texture
	if err := json.NewDecoder(resp.Body).Decode(&textures); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return textures, nil
}

func (c *Client) CreateOrder(ctx context.Context, req OrderRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/api/orders", c.baseURL),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetTexturePrice(ctx context.Context, textureID string) (float64, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/api/textures/%s/price", c.baseURL, textureID),
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Price float64 `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	return result.Price, nil
}
