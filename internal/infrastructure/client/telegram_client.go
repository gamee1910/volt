package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gamee1910/volt/config"
	"github.com/gamee1910/volt/internal/domain/ports"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/internal/interfaces/api/handler/request"
	"github.com/gamee1910/volt/pkg/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type TelegramClient struct {
	bot                *bot.Bot
	cfg                *config.Configuration
	log                *logger.Logger
	electricityService service.ElectricityService
}

func NewTelegramClient(
	cfg *config.Configuration,
	log *logger.Logger,
	electricityService service.ElectricityService,
) (ports.TelegramClient, error) {
	tc := &TelegramClient{
		cfg:                cfg,
		log:                log,
		electricityService: electricityService,
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(tc.defaultHandler()),
	}

	b, err := bot.New(cfg.ApplicationConfig.TelegramConfig.TelegramAPIKey, opts...)
	if err != nil {
		log.Error("failed_to_create_telegram_bot", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	tc.bot = b
	return tc, nil
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
		c.log.Error("failed_to_send_telegram_message", map[string]interface{}{
			"chat_id": chatID,
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	return nil
}

func (c *TelegramClient) handleYesterdayCommand(ctx context.Context, chatID int64) error {
	usage, err := c.electricityService.GetYesterDayUsage(ctx)
	if err != nil {
		c.log.Error("failed_to_get_yesterday_usage", map[string]interface{}{
			"error": err.Error(),
		})
		return c.SendMessage(ctx, chatID, "Failed to fetch data: "+err.Error())
	}

	dateStr := usage.MeasurementDate.Format("02/01/2006")
	kwhStr := fmt.Sprintf("%.2f KWh", usage.ConsumptionKWh)
	amountStr := formatVND(usage.TotalAmount)

	message := fmt.Sprintf(
		"Điện năng ngày %s:\n Tiêu thụ: %s\n Tổng tiền tháng này: %s",
		dateStr,
		kwhStr,
		amountStr,
	)

	return c.SendMessage(ctx, chatID, message)
}

func (c *TelegramClient) handleLogin(ctx context.Context, chatID int64) error {
	err := c.electricityService.LoginEVN(ctx, c.cfg.ApplicationConfig.EnvConfig.Username, c.cfg.ApplicationConfig.EnvConfig.Password)
	if err != nil {
		c.log.Error("failed_to_login", map[string]interface{}{
			"error": err.Error(),
		})
		return c.SendMessage(ctx, chatID, "Đăng nhập thất bại: "+err.Error())
	}

	return c.SendMessage(ctx, chatID, "Đăng nhập thành công")
}

func (c *TelegramClient) handleGetAllCommand(ctx context.Context, chatID int64) error {
	resp, err := c.electricityService.GetAll(ctx)
	if err != nil {
		c.log.Error("failed_to_get_all_usage", map[string]interface{}{
			"error": err.Error(),
		})
		return c.SendMessage(ctx, chatID, "Failed to fetch data: "+err.Error())
	}

	if len(resp.Data) == 0 {
		return c.SendMessage(ctx, chatID, "Chưa có dữ liệu sản lượng điện.")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Danh sách sản lượng điện (%d ngày):\n\n", len(resp.Data)))

	for _, usage := range resp.Data {
		dateStr := usage.MeasurementDate.Format("02/01/2006")
		kwhStr := fmt.Sprintf("%.2f KWh", usage.ConsumptionKWh)
		amountStr := formatVND(usage.TotalAmount)
		sb.WriteString(fmt.Sprintf("• Ngày %s: %s | %s\n", dateStr, kwhStr, amountStr))
	}

	sb.WriteString(fmt.Sprintf("\n Tổng tiêu thụ: %.2f KWh\n Tổng tiền ước tính: %s", resp.TotalKWh, formatVND(resp.TotalAmount)))

	messageText := sb.String()
	if len(messageText) > 4000 {
		return c.sendChunkedMessages(ctx, chatID, messageText)
	}

	return c.SendMessage(ctx, chatID, messageText)
}

func (c *TelegramClient) sendChunkedMessages(ctx context.Context, chatID int64, fullText string) error {
	lines := strings.Split(fullText, "\n")
	var currentChunk strings.Builder

	for _, line := range lines {
		if currentChunk.Len()+len(line)+1 > 4000 {
			if err := c.SendMessage(ctx, chatID, currentChunk.String()); err != nil {
				return err
			}
			currentChunk.Reset()
		}
		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")
	}

	if currentChunk.Len() > 0 {
		return c.SendMessage(ctx, chatID, currentChunk.String())
	}
	return nil
}

func (c *TelegramClient) handleSyncCommand(ctx context.Context, chatID int64, text string) error {
	parts := strings.Fields(text)
	var fromDate, toDate string

	if len(parts) >= 3 {
		fromDate = parts[1]
		toDate = parts[2]
	} else {
		loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
		if err != nil {
			loc = time.Local
		}
		now := time.Now().In(loc)
		fromDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).Format("02/01/2006")
		toDate = now.Format("02/01/2006")
	}

	req := request.DailyPowerUsageRequest{
		Token:        "",
		CustomerCode: c.cfg.ApplicationConfig.EnvConfig.CustomerCode,
		FromDate:     fromDate,
		ToDate:       toDate,
	}

	if err := c.electricityService.FetchAndSyncMonthlyUsage(ctx, req); err != nil {
		c.log.Error("failed_to_sync_evn_data", map[string]interface{}{
			"from_date": fromDate,
			"to_date":   toDate,
			"error":     err.Error(),
		})
		return c.SendMessage(ctx, chatID, fmt.Sprintf("Sync failed: %s", err.Error()))
	}

	return c.SendMessage(ctx, chatID, fmt.Sprintf("Đồng bộ dữ liệu thành công từ %s đến %s!", fromDate, toDate))
}

func (c *TelegramClient) defaultHandler() func(ctx context.Context, b *bot.Bot, update *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}

		text := update.Message.Text
		chatID := update.Message.Chat.ID

		switch parseCommand(text) {
		case "/yesterday":
			if err := c.handleYesterdayCommand(ctx, chatID); err != nil {
				c.log.Error("handle_yesterday_failed", map[string]interface{}{"error": err.Error()})
			}
		case "/login":
			if err := c.handleLogin(ctx, chatID); err != nil {
				c.log.Error("handle_login_failed", map[string]interface{}{"error": err.Error()})
			}
		case "/sync":
			if err := c.handleSyncCommand(ctx, chatID, text); err != nil {
				c.log.Error("handle_sync_failed", map[string]interface{}{"error": err.Error()})
			}
		case "/get":
			if err := c.handleGetAllCommand(ctx, chatID); err != nil {
				c.log.Error("handle_get_failed", map[string]interface{}{"error": err.Error()})
			}
		case "/start", "/help":
			msg := "Volt Telegram Bot\n\nDanh sách lệnh:\n/yesterday - Xem sản lượng điện ngày hôm qua\n/sync [from_date] [to_date] - Đồng bộ dữ liệu từ EVN (mặc định: từ đầu tháng)"
			if err := c.SendMessage(ctx, chatID, msg); err != nil {
				c.log.Error("handle_help_failed", map[string]interface{}{"error": err.Error()})
			}
		}
	}
}

func parseCommand(text string) string {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return ""
	}
	cmd := parts[0]
	cmd = strings.Split(cmd, "@")[0]
	return cmd
}

func formatVND(amount float64) string {
	p := message.NewPrinter(language.Vietnamese)
	return p.Sprintf("%.0f VND", amount)
}
