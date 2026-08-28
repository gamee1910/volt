package worker

import (
	"context"
	"time"

	"github.com/gamee1910/volt/internal/application/types"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/pkg/logger"
	"github.com/gamee1910/volt/pkg/utils"
)

type ElectricityWorker struct {
	electricityService service.ElectricityService
	log                *logger.Logger
}

func NewElectricityWorker(
	electricityService service.ElectricityService,
	log *logger.Logger,
) *ElectricityWorker {
	return &ElectricityWorker{
		electricityService: electricityService,
		log:                log,
	}
}

func (w *ElectricityWorker) Run(ctx context.Context) error {
	location, err := utils.LoadVietnamTimezone()
	if err != nil {
		return err
	}

	now := time.Now().In(location)
	firstDayOfMonth := utils.GetFirstDayOfMonth(now, location)

	return w.electricityService.DailyPowerUsage(ctx, types.GetUsageParam{
		FromDate: utils.FormatEVNDate(firstDayOfMonth),
		ToDate:   utils.FormatEVNDate(now),
	})
}
