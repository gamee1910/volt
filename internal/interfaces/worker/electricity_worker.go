package worker

import (
	"context"
	"time"

	"github.com/gamee1910/volt/internal/application"
	"github.com/gamee1910/volt/internal/application/command"
	"github.com/gamee1910/volt/internal/common/utils"
	"github.com/gamee1910/volt/pkg/logger"
)

type ElectricityWorker struct {
	app *application.Application
	log *logger.Logger
}

func NewElectricityWorker(
	app *application.Application,
	log *logger.Logger,
) *ElectricityWorker {
	return &ElectricityWorker{
		app: app,
		log: log,
	}
}

func (w *ElectricityWorker) Run(ctx context.Context) error {
	location, err := utils.LoadVietnamTimezone()
	if err != nil {
		return err
	}

	now := time.Now().In(location)
	firstDayOfMonth := utils.GetFirstDayOfMonth(now, location)

	return w.app.Commands.SyncElectricity.Handle(ctx, command.SyncElectricityCommand{
		FromDate: utils.FormatEVNDate(firstDayOfMonth),
		ToDate:   utils.FormatEVNDate(now),
	})
}
