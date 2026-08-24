package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gamee1910/volt/config"
	"github.com/gamee1910/volt/internal/application/dto"
	"github.com/gamee1910/volt/internal/application/port"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/pkg/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type ElectricityHandler struct {
	cfg                *config.Configuration
	log                *logger.Logger
	sender             port.TelegramClient
	electricityService service.ElectricityService
}

func NewElectricityHandler(
	cfg *config.Configuration,
	log *logger.Logger,
	sender port.TelegramClient,
	electricityService service.ElectricityService,
) *ElectricityHandler {
	return &ElectricityHandler{
		cfg:                cfg,
		log:                log,
		sender:             sender,
		electricityService: electricityService,
	}
}

func (h *ElectricityHandler) SetSender(sender port.TelegramClient) {
	h.sender = sender
}

func (h *ElectricityHandler) DefaultHandler() func(ctx context.Context, b *bot.Bot, update *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}

		text := update.Message.Text
		chatID := update.Message.Chat.ID

		switch parseCommand(text) {
		case "/yesterday":
			if err := h.handleYesterday(ctx, chatID); err != nil {
				h.log.Error("handle_yesterday_failed", map[string]interface{}{"error": err.Error()})
			}
		case "/login":
			if err := h.handleLogin(ctx, chatID); err != nil {
				h.log.Error("handle_login_failed", map[string]interface{}{"error": err.Error()})
			}
		case "/sync":
			if err := h.handleSync(ctx, chatID, text); err != nil {
				h.log.Error("handle_sync_failed", map[string]interface{}{"error": err.Error()})
			}
		case "/get":
			if err := h.handleGetAll(ctx, chatID); err != nil {
				h.log.Error("handle_get_failed", map[string]interface{}{"error": err.Error()})
			}
		case "/start", "/help":
			msg := "Volt Telegram Bot\n\nDanh sách lệnh:\n/yesterday - Xem sản lượng điện ngày hôm qua\n/sync [from_date] [to_date] - Đồng bộ dữ liệu từ EVN (mặc định: từ đầu tháng)\n/get - Xem toàn bộ dữ liệu"
			if err := h.sender.SendMessage(ctx, chatID, msg); err != nil {
				h.log.Error("handle_help_failed", map[string]interface{}{"error": err.Error()})
			}
		}
	}
}

func (h *ElectricityHandler) handleYesterday(ctx context.Context, chatID int64) error {
	usage, err := h.electricityService.GetYesterDayUsage(ctx)
	if err != nil {
		h.log.Error("failed_to_get_yesterday_usage", map[string]interface{}{"error": err.Error()})
		return h.sender.SendMessage(ctx, chatID, "Failed to fetch data: "+err.Error())
	}

	msg := fmt.Sprintf(
		"Điện năng ngày %s:\n Tiêu thụ: %.2f KWh\n Tổng tiền tháng này: %s",
		usage.MeasurementDate.Format("02/01/2006"),
		usage.ConsumptionKWh,
		formatVND(usage.TotalAmount),
	)
	return h.sender.SendMessage(ctx, chatID, msg)
}

func (h *ElectricityHandler) handleLogin(ctx context.Context, chatID int64) error {
	err := h.electricityService.LoginEVN(
		ctx,
		h.cfg.ApplicationConfig.EnvConfig.Username,
		h.cfg.ApplicationConfig.EnvConfig.Password,
	)
	if err != nil {
		h.log.Error("failed_to_login", map[string]interface{}{"error": err.Error()})
		return h.sender.SendMessage(ctx, chatID, "Đăng nhập thất bại: "+err.Error())
	}
	return h.sender.SendMessage(ctx, chatID, "Đăng nhập thành công")
}

func (h *ElectricityHandler) handleGetAll(ctx context.Context, chatID int64) error {
	resp, err := h.electricityService.GetAll(ctx)
	if err != nil {
		h.log.Error("failed_to_get_all_usage", map[string]interface{}{"error": err.Error()})
		return h.sender.SendMessage(ctx, chatID, "Failed to fetch data: "+err.Error())
	}

	if len(resp.Data) == 0 {
		return h.sender.SendMessage(ctx, chatID, "Chưa có dữ liệu sản lượng điện.")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Danh sách sản lượng điện (%d ngày):\n\n", len(resp.Data)))
	for _, usage := range resp.Data {
		sb.WriteString(fmt.Sprintf("• Ngày %s: %.2f KWh | %s\n",
			usage.MeasurementDate.Format("02/01/2006"),
			usage.ConsumptionKWh,
			formatVND(usage.TotalAmount),
		))
	}
	sb.WriteString(fmt.Sprintf("\n Tổng tiêu thụ: %.2f KWh\n Tổng tiền ước tính: %s",
		resp.TotalKWh, formatVND(resp.TotalAmount),
	))

	msgText := sb.String()
	if len(msgText) > 4000 {
		return h.sendChunked(ctx, chatID, msgText)
	}
	return h.sender.SendMessage(ctx, chatID, msgText)
}

func (h *ElectricityHandler) handleSync(ctx context.Context, chatID int64, text string) error {
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

	req := dto.DailyPowerUsageRequest{
		Token:        "",
		CustomerCode: h.cfg.ApplicationConfig.EnvConfig.CustomerCode,
		FromDate:     fromDate,
		ToDate:       toDate,
	}

	if err := h.electricityService.FetchAndSyncMonthlyUsage(ctx, req); err != nil {
		h.log.Error("failed_to_sync_evn_data", map[string]interface{}{
			"from_date": fromDate,
			"to_date":   toDate,
			"error":     err.Error(),
		})
		return h.sender.SendMessage(ctx, chatID, fmt.Sprintf("Sync failed: %s", err.Error()))
	}

	return h.sender.SendMessage(ctx, chatID, fmt.Sprintf("Đồng bộ dữ liệu thành công từ %s đến %s!", fromDate, toDate))
}

func (h *ElectricityHandler) sendChunked(ctx context.Context, chatID int64, fullText string) error {
	lines := strings.Split(fullText, "\n")
	var currentChunk strings.Builder

	for _, line := range lines {
		if currentChunk.Len()+len(line)+1 > 4000 {
			if err := h.sender.SendMessage(ctx, chatID, currentChunk.String()); err != nil {
				return err
			}
			currentChunk.Reset()
		}
		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")
	}

	if currentChunk.Len() > 0 {
		return h.sender.SendMessage(ctx, chatID, currentChunk.String())
	}
	return nil
}

func parseCommand(text string) string {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return ""
	}
	return strings.Split(parts[0], "@")[0]
}

func formatVND(amount float64) string {
	p := message.NewPrinter(language.Vietnamese)
	return p.Sprintf("%.0f VND", amount)
}
