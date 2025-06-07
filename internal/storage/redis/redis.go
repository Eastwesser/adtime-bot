package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

const stateTTL = 24 * time.Hour

type Storage struct {
	client *redis.Client
}

// New creates a new Redis client
func New(addr, password string, db int) *Storage {
	return &Storage{
		client: redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     password,
			DB:           db,
			PoolSize:     100, // Increase connection pool size
			MinIdleConns: 10,  // Keep minimum connections ready
		}),
	}
}

// Close closes the Redis connection
func (s *Storage) Close() {
	if s.client != nil {
		_ = s.client.Close()
	}
}

func (s *Storage) SetUserDialogState(ctx context.Context, chatId int64, state *UserState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	return s.client.Set(ctx, buildStateKey(chatId), data, stateTTL).Err()
}

func (s *Storage) GetUserDialogState(ctx context.Context, chatID int64) (*UserState, error) {
	data, err := s.client.Get(ctx, buildStateKey(chatID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return &UserState{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get state: %w", err)
	}

	var state UserState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal failure: %w", err)
	}
	return &state, nil
}

func (s *Storage) DropUserDialogState(ctx context.Context, chatID int64) error {
	return s.client.Del(ctx, buildStateKey(chatID)).Err()
}

func buildStateKey(chatId int64) string {
	return fmt.Sprintf("state:%d", chatId)
}
