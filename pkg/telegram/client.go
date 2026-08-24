package telegram

import (
	"context"
	"fmt"

	"github.com/gamee1910/volt/pkg/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type MessageHandler func(
	ctx context.Context, chatID int64, text string,
)

type ElectricityTelegramClient struct {
	bot            *bot.Bot
	log            *logger.Logger
	messageHandler MessageHandler
}

func NewTelegramClient(
	apiKey string,
	log *logger.Logger,
) (TelegramClient, error) {
	client := &ElectricityTelegramClient{
		log: log,
	}

	handler := func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		if client.messageHandler == nil {
			return
		}
		client.messageHandler(ctx, update.Message.Chat.ID, update.Message.Text)
	}

	options := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(apiKey, options...)
	if err != nil {
		log.Error("failed_to_create_telegram_bot", map[string]any{
			"error": err.Error(),
		})

		return nil, fmt.Errorf(
			"failed to create telegram bot: %w",
			err,
		)
	}

	client.bot = b

	return client, nil
}

func (c *ElectricityTelegramClient) Start(ctx context.Context) error {
	c.log.Info("telegram_bot_started")
	go c.bot.Start(ctx)
	return nil
}

func (c *ElectricityTelegramClient) SendMessage(
	ctx context.Context, chatID int64, text string,
) error {
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

func (c *ElectricityTelegramClient) SetMessageHandler(
	handler MessageHandler,
) {
	c.messageHandler = handler
}
