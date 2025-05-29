package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CLIENT'S STATE IN REDIS

type Client struct {
	client *redis.Client
	ttl    time.Duration
}

func (c *Client) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	return c.client.Get(ctx, key).Bytes()
}

func (c *Client) Set(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *Client) Close() {
	if c.client != nil {
		_ = c.client.Close()
	}
}

func New(addr, password string, db int, ttl time.Duration) *Client {
	return &Client{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		ttl: ttl,
	}
}

func (c *Client) SaveState(ctx context.Context, chatID int64, state any) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	return c.client.Set(ctx, fmt.Sprintf("state:%d", chatID), data, c.ttl).Err()
}

func (c *Client) GetState(ctx context.Context, chatID int64, state interface{}) error {
	data, err := c.client.Get(ctx, fmt.Sprintf("state:%d", chatID)).Bytes()
	if err != nil {
		return fmt.Errorf("get state: %w", err)
	}

	return json.Unmarshal(data, state)
}

func (c *Client) ClearState(ctx context.Context, chatID int64) error {
	return c.client.Del(ctx, fmt.Sprintf("state:%d", chatID)).Err()
}
