package start

import "context"

type StateManager interface {
	SetStep(ctx context.Context, chatID int64, step string) error
	ResetDialogState(ctx context.Context, chatID int64) error
}

type PgStorage interface {
	GetUserAgreement(ctx context.Context, userID int64) (bool, string, error)
}
