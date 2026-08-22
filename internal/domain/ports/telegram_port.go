package ports

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramClient interface {
	SendMessage(ctx context.Context, b *bot.Bot, update *models.Update)
}
