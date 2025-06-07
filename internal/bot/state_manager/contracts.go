package state_manager

import (
	"adtime-bot/internal/storage/redis"
	"context"
)

type RedisStorage interface {
	GetUserDialogState(ctx context.Context, chatId int64) (*redis.UserState, error)
	SetUserDialogState(ctx context.Context, chatId int64, state *redis.UserState) error
	DropUserDialogState(ctx context.Context, chatId int64) error
}

var _ RedisStorage = (*redis.Storage)(nil)
