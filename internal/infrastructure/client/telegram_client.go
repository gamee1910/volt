package client

import (
	"context"
	"fmt"

	"github.com/gamee1910/volt/internal/application/port"
	"github.com/gamee1910/volt/pkg/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramClient struct {
	bot *bot.Bot
	log *logger.Logger
}

func NewTelegramClient(
	apiKey string,
	log *logger.Logger,
	defaultHandler func(ctx context.Context, b *bot.Bot, update *models.Update),
) (port.TelegramClient, error) {
	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}

	b, err := bot.New(apiKey, opts...)
	if err != nil {
		log.Error("failed_to_create_telegram_bot", map[string]any{
			"error": err.Error(),
		})
		return nil, err
	}

	return &TelegramClient{bot: b, log: log}, nil
}

func (c *TelegramClient) Start(ctx context.Context) error {
	c.log.Info("telegram_bot_started")
	go c.bot.Start(ctx)
	return nil
}

func (c *TelegramClient) SendMessage(ctx context.Context, chatID int64, text string) error {
	_, err := c.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		c.log.Error("failed_to_send_telegram_message", map[string]any{
			"chat_id": chatID,
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	return nil
}
