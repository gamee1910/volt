package query

import (
	"context"
	"fmt"
	"time"

	"github.com/gamee1910/volt/internal/application/types"
	"github.com/gamee1910/volt/internal/common/decorator"
	"github.com/gamee1910/volt/internal/common/utils"
	"github.com/gamee1910/volt/internal/domain/repository"
	"github.com/gamee1910/volt/pkg/logger"
)

type YesterdayUsageQuery struct{}

type YesterdayHandler decorator.QueryHandler[
	YesterdayUsageQuery,
	types.ElectricityConsumption,
]

type yesterdayHandler struct {
	repo repository.ElectricityRepository
}

func NewYesterdayUsageHandler(
	repo repository.ElectricityRepository,
	logger *logger.Logger,
	metricClient decorator.MetricClient,
) YesterdayHandler {
	return decorator.ApplyQueryDecorators(
		yesterdayHandler{repo: repo}, logger, metricClient,
	)
}

func (h yesterdayHandler) Handle(
	ctx context.Context,
	_ YesterdayUsageQuery,
) (electricity types.ElectricityConsumption, err error) {
	location, err := utils.LoadVietnamTimezone()
	if err != nil {
		return types.ElectricityConsumption{}, fmt.Errorf("failed to load viet nam timezone")
	}

	now := time.Now().In(location)
	yesterday := utils.GetYesterday(now, location)
	firstDayOfMonth := utils.GetFirstDayOfMonth(now, location)

	consumption, err := h.repo.FetchByDate(ctx, yesterday)
	if err != nil {
		return types.ElectricityConsumption{}, fmt.Errorf("failed to fetch yesterday usage: %w", err)
	}

	totalKWh, err := h.repo.CalculateTotalConsumptionFromDateToDate(
		ctx,
		firstDayOfMonth,
		yesterday,
	)
	if err != nil {
		return types.ElectricityConsumption{}, fmt.Errorf("failed to calculate total consumption: %w", err)
	}

	totalAmount := utils.CalculateElectricityBill(totalKWh)

	return types.ElectricityConsumption{
		MeasurementDate: consumption.MeasurementDate,
		ConsumptionKWh:  consumption.ConsumptionKWh,
		TotalAmount:     totalAmount,
	}, nil
}
