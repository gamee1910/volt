package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gamee1910/volt/config"
	"github.com/gamee1910/volt/internal/application"
	"github.com/gamee1910/volt/internal/application/command"
	"github.com/gamee1910/volt/internal/application/query"
	"github.com/gamee1910/volt/internal/common/utils"
	"github.com/gamee1910/volt/pkg/logger"
	"github.com/gamee1910/volt/pkg/telegram"
)

type ElectricityCommand struct {
	cfg            *config.Configuration
	log            *logger.Logger
	telegramClient telegram.TelegramClient
	app            *application.Application
}

func NewElectricityCommand(
	cfg *config.Configuration,
	log *logger.Logger,
	telegramClient telegram.TelegramClient,
	app *application.Application,
) *ElectricityCommand {
	return &ElectricityCommand{
		cfg:            cfg,
		log:            log,
		telegramClient: telegramClient,
		app:            app,
	}
}

func (c *ElectricityCommand) HandleYesterdayCommand(ctx context.Context, chatID int64) error {
	usage, err := c.app.Queries.YesterdayUsage.Handle(ctx, query.YesterdayUsageQuery{})
	if err != nil {
		c.log.Error("failed_to_get_yesterday_usage", map[string]any{"error": err.Error()})
		return c.telegramClient.SendMessage(ctx, chatID, "failed to fetch data: "+err.Error())
	}

	msg := fmt.Sprintf(
		"Điện năng ngày %s:\n Tiêu thụ: %.2f KWh\n Tổng tiền tháng này: %s",
		usage.MeasurementDate.Format("02/01/2006"),
		usage.ConsumptionKWh,
		utils.FormatVND(usage.TotalAmount),
	)
	return c.telegramClient.SendMessage(ctx, chatID, msg)
}

func (c *ElectricityCommand) HandleGetAllCommand(ctx context.Context, chatID int64) error {
	resp, err := c.app.Queries.AllElectricity.Handle(ctx, query.AllElectricityQuery{})
	if err != nil {
		c.log.Error("failed_to_get_all_usage", map[string]any{"error": err.Error()})
		return c.telegramClient.SendMessage(ctx, chatID, "Failed to fetch data: "+err.Error())
	}

	if len(resp.Data) == 0 {
		return c.telegramClient.SendMessage(ctx, chatID, "Chưa có dữ liệu sản lượng điện.")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Danh sách sản lượng điện (%d ngày):\n\n", len(resp.Data)))
	for _, usage := range resp.Data {
		sb.WriteString(fmt.Sprintf("• Ngày %s: %.2f KWh | %s\n",
			usage.MeasurementDate.Format("02/01/2006"),
			usage.ConsumptionKWh,
			utils.FormatVND(usage.TotalAmount),
		))
	}
	sb.WriteString(fmt.Sprintf("\n Tổng tiêu thụ: %.2f KWh\n Tổng tiền ước tính: %s",
		resp.TotalKWh, utils.FormatVND(resp.TotalAmount),
	))

	msgText := sb.String()
	if len(msgText) > 4000 {
		return c.sendChunked(ctx, chatID, msgText)
	}
	return c.telegramClient.SendMessage(ctx, chatID, msgText)
}

func (c *ElectricityCommand) HandleSyncCommand(ctx context.Context, chatID int64, text string) error {
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

	if err := c.app.Commands.SyncElectricity.Handle(ctx, command.SyncElectricityCommand{
		FromDate: fromDate,
		ToDate:   toDate,
	}); err != nil {
		c.log.Error("failed_to_sync_evn_data", map[string]any{
			"from_date": fromDate,
			"to_date":   toDate,
			"error":     err.Error(),
		})
		return c.telegramClient.SendMessage(ctx, chatID, fmt.Sprintf("Sync failed: %s", err.Error()))
	}

	return c.telegramClient.SendMessage(ctx, chatID, fmt.Sprintf("Đồng bộ dữ liệu thành công từ %s đến %s!", fromDate, toDate))
}

func (c *ElectricityCommand) sendChunked(ctx context.Context, chatID int64, fullText string) error {
	lines := strings.Split(fullText, "\n")
	var currentChunk strings.Builder

	for _, line := range lines {
		if currentChunk.Len()+len(line)+1 > 4000 {
			if err := c.telegramClient.SendMessage(ctx, chatID, currentChunk.String()); err != nil {
				return err
			}
			currentChunk.Reset()
		}
		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")
	}

	if currentChunk.Len() > 0 {
		return c.telegramClient.SendMessage(ctx, chatID, currentChunk.String())
	}
	return nil
}
