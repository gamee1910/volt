package router

import (
	"context"
	"strings"

	"github.com/gamee1910/volt/internal/application/port"
	"github.com/gamee1910/volt/internal/infrastructure/client"
	"github.com/gamee1910/volt/internal/interfaces/bot/command"
	"github.com/gamee1910/volt/pkg/logger"
)

type Router struct {
	log         *logger.Logger
	sender      port.TelegramClient
	electricity *command.ElectricityCommand
}

func NewRouter(
	log *logger.Logger,
	sender port.TelegramClient,
	electricity *command.ElectricityCommand,
) *Router {
	return &Router{log: log, sender: sender, electricity: electricity}
}

func (r *Router) SetSender(sender port.TelegramClient) {
	r.sender = sender
}

// DefaultHandler trả về MessageHandler — type thuần Go, không phụ thuộc go-telegram/bot
func (r *Router) DefaultHandler() client.MessageHandler {
	return func(ctx context.Context, chatID int64, text string) {
		switch parseCommand(text) {
		case "/yesterday":
			if err := r.electricity.Yesterday(ctx, chatID); err != nil {
				r.log.Error("handle_yesterday_failed", map[string]any{"error": err.Error()})
			}
		case "/login":
			if err := r.electricity.Login(ctx, chatID); err != nil {
				r.log.Error("handle_login_failed", map[string]any{"error": err.Error()})
			}
		case "/sync":
			if err := r.electricity.Sync(ctx, chatID, text); err != nil {
				r.log.Error("handle_sync_failed", map[string]any{"error": err.Error()})
			}
		case "/get":
			if err := r.electricity.GetAll(ctx, chatID); err != nil {
				r.log.Error("handle_get_failed", map[string]any{"error": err.Error()})
			}
		case "/start", "/help":
			msg := "Volt Telegram Bot\n\nDanh sách lệnh:\n/yesterday - Xem sản lượng điện ngày hôm qua\n/sync [from_date] [to_date] - Đồng bộ dữ liệu từ EVN (mặc định: từ đầu tháng)\n/get - Xem toàn bộ dữ liệu"
			if err := r.sender.SendMessage(ctx, chatID, msg); err != nil {
				r.log.Error("handle_help_failed", map[string]any{"error": err.Error()})
			}
		}
	}
}

func parseCommand(text string) string {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return ""
	}
	return strings.Split(parts[0], "@")[0]
}
