package command

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gamee1910/volt/internal/common/decorator"
	"github.com/gamee1910/volt/internal/common/utils"
	"github.com/gamee1910/volt/internal/domain/entity"
	"github.com/gamee1910/volt/internal/domain/repository"
	"github.com/gamee1910/volt/pkg/evnhcmc"
	"github.com/gamee1910/volt/pkg/logger"
)

type SyncElectricityCommand struct {
	FromDate string
	ToDate   string
}

type SyncElectricityHandler decorator.CommandHandler[SyncElectricityCommand]

type syncElectricityHandler struct {
	repo      repository.ElectricityRepository
	evnClient evnhcmc.EVNClient
}

func NewSyncElectricityHandler(
	repo repository.ElectricityRepository,
	evnClient evnhcmc.EVNClient,
	logger *logger.Logger,
	metricClient decorator.MetricClient,
) decorator.CommandHandler[SyncElectricityCommand] {
	return decorator.ApplyCommandDecorators(
		syncElectricityHandler{
			repo:      repo,
			evnClient: evnClient,
		},
		logger,
		metricClient,
	)
}

func (h syncElectricityHandler) Handle(ctx context.Context, cmd SyncElectricityCommand) (err error) {
	resp, err := h.evnClient.GetDailyPowerUsageData(ctx, evnhcmc.DailyPowerUsageRequest{
		FromDate: cmd.FromDate,
		ToDate:   cmd.ToDate,
	})

	if err != nil {
		return fmt.Errorf("failed to fetch EVN data: %w", err)
	}

	for _, item := range resp.Data.DailyOutputs {
		kwh, err := strconv.ParseFloat(item.TotalOutput, 64)
		if err != nil {
			return fmt.Errorf(
				"invalid total output %q at date %s: %w",
				item.TotalOutput,
				item.FullDate,
				err,
			)
		}

		if kwh == 0 {
			kwh = item.TotalIndex
		}

		location, err := utils.LoadVietnamTimezone()
		if err != nil {
			return fmt.Errorf("failed to load Viet Nam timezone: %w", err)
		}

		measurementDate, err := utils.ParseEVNDate(item.FullDate, location)
		if err != nil {
			return fmt.Errorf("failed to parse evn date: %q", item.FullDate)
		}

		consumption := &entity.ElectricityConsumption{
			MeasurementDate: measurementDate,
			ConsumptionKWh:  kwh,
		}

		if err := h.repo.Upsert(ctx, consumption); err != nil {
			return fmt.Errorf(
				"failed to upsert consumption at date: %q: %w",
				item.FullDate,
				err,
			)
		}
	}
	return nil
}
