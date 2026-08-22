package ports

import (
	"context"
)

type TelegramClient interface {
	Start(ctx context.Context) error
	SendMessage(ctx context.Context, chatID int64, text string) error
}
