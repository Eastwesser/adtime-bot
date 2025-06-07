package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StateManager interface {
}

type CommandHandler interface {
	Handle(ctx context.Context, msg *tgbotapi.Message)
}

type CallbackHandler interface {
}

type MessageHandlers interface {
}

type ButtonHandlers interface {
}
