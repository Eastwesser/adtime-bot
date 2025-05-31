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

// Expire sets a key's time to live (TTL)
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.client.Expire(ctx, key, expiration).Result()
}

// Incr increments the key's value by 1. Returns the new value and any error
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// Del deletes a key
func (c *Client) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Get retrieves a key's value
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	return c.client.Get(ctx, key).Bytes()
}

// Set sets a key's value with TTL
func (c *Client) Set(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, data, ttl).Err()
}

// Close closes the Redis connection
func (c *Client) Close() {
	if c.client != nil {
		_ = c.client.Close()
	}
}

// New creates a new Redis client
func New(addr, password string, db int, ttl time.Duration) *Client {
	return &Client{
		client: redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     password,
			DB:           db,
			PoolSize:     100, // Increase connection pool size
			MinIdleConns: 10,  // Keep minimum connections ready
		}),
		ttl: ttl,
	}
}

// SaveState saves user state to Redis
func (c *Client) SaveState(ctx context.Context, chatID int64, state any) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	return c.client.Set(ctx, fmt.Sprintf("state:%d", chatID), data, c.ttl).Err()
}

// GetState retrieves user state from Redis
func (c *Client) GetState(ctx context.Context, chatID int64, state interface{}) error {
	data, err := c.client.Get(ctx, fmt.Sprintf("state:%d", chatID)).Bytes()
	if err != nil {
		return fmt.Errorf("get state: %w", err)
	}

	return json.Unmarshal(data, state)
}

// ClearState removes user state from Redis
func (c *Client) ClearState(ctx context.Context, chatID int64) error {
	return c.client.Del(ctx, fmt.Sprintf("state:%d", chatID)).Err()
}
