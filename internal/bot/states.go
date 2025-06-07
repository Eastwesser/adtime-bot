package bot

import (
	"adtime-bot/internal/storage"
	"adtime-bot/pkg/redis"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type StateStorage struct {
	ttl time.Duration
}

func NewStateStorage() *StateStorage {
	return &StateStorage{
		ttl: 24 * time.Hour,
	}
}
