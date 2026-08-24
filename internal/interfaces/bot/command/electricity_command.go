package command

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
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type ElectricityCommand struct {
	cfg                *config.Configuration
	log                *logger.Logger
	sender             port.TelegramClient
	electricityService service.ElectricityService
}

func NewElectricityCommand(
	cfg *config.Configuration,
	log *logger.Logger,
	sender port.TelegramClient,
	electricityService service.ElectricityService,
) *ElectricityCommand {
	return &ElectricityCommand{
		cfg:                cfg,
		log:                log,
		sender:             sender,
		electricityService: electricityService,
	}
}

func (c *ElectricityCommand) SetSender(sender port.TelegramClient) {
	c.sender = sender
}

func (c *ElectricityCommand) Yesterday(ctx context.Context, chatID int64) error {
	usage, err := c.electricityService.GetYesterDayUsage(ctx)
	if err != nil {
		c.log.Error("failed_to_get_yesterday_usage", map[string]any{"error": err.Error()})
		return c.sender.SendMessage(ctx, chatID, "Failed to fetch data: "+err.Error())
	}

	msg := fmt.Sprintf(
		"Điện năng ngày %s:\n Tiêu thụ: %.2f KWh\n Tổng tiền tháng này: %s",
		usage.MeasurementDate.Format("02/01/2006"),
		usage.ConsumptionKWh,
		formatVND(usage.TotalAmount),
	)
	return c.sender.SendMessage(ctx, chatID, msg)
}

func (c *ElectricityCommand) Login(ctx context.Context, chatID int64) error {
	err := c.electricityService.LoginEVN(
		ctx,
		c.cfg.ApplicationConfig.EnvConfig.Username,
		c.cfg.ApplicationConfig.EnvConfig.Password,
	)
	if err != nil {
		c.log.Error("failed_to_login", map[string]any{"error": err.Error()})
		return c.sender.SendMessage(ctx, chatID, "Đăng nhập thất bại: "+err.Error())
	}
	return c.sender.SendMessage(ctx, chatID, "Đăng nhập thành công")
}

func (c *ElectricityCommand) GetAll(ctx context.Context, chatID int64) error {
	resp, err := c.electricityService.GetAll(ctx)
	if err != nil {
		c.log.Error("failed_to_get_all_usage", map[string]any{"error": err.Error()})
		return c.sender.SendMessage(ctx, chatID, "Failed to fetch data: "+err.Error())
	}

	if len(resp.Data) == 0 {
		return c.sender.SendMessage(ctx, chatID, "Chưa có dữ liệu sản lượng điện.")
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
		return c.sendChunked(ctx, chatID, msgText)
	}
	return c.sender.SendMessage(ctx, chatID, msgText)
}

func (c *ElectricityCommand) Sync(ctx context.Context, chatID int64, text string) error {
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
		CustomerCode: c.cfg.ApplicationConfig.EnvConfig.CustomerCode,
		FromDate:     fromDate,
		ToDate:       toDate,
	}

	if err := c.electricityService.FetchAndSyncMonthlyUsage(ctx, req); err != nil {
		c.log.Error("failed_to_sync_evn_data", map[string]any{
			"from_date": fromDate,
			"to_date":   toDate,
			"error":     err.Error(),
		})
		return c.sender.SendMessage(ctx, chatID, fmt.Sprintf("Sync failed: %s", err.Error()))
	}

	return c.sender.SendMessage(ctx, chatID, fmt.Sprintf("Đồng bộ dữ liệu thành công từ %s đến %s!", fromDate, toDate))
}

func (c *ElectricityCommand) sendChunked(ctx context.Context, chatID int64, fullText string) error {
	lines := strings.Split(fullText, "\n")
	var currentChunk strings.Builder

	for _, line := range lines {
		if currentChunk.Len()+len(line)+1 > 4000 {
			if err := c.sender.SendMessage(ctx, chatID, currentChunk.String()); err != nil {
				return err
			}
			currentChunk.Reset()
		}
		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")
	}

	if currentChunk.Len() > 0 {
		return c.sender.SendMessage(ctx, chatID, currentChunk.String())
	}
	return nil
}

func formatVND(amount float64) string {
	p := message.NewPrinter(language.Vietnamese)
	return p.Sprintf("%.0f VND", amount)
}
