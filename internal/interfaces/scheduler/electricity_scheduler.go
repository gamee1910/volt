package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/gamee1910/volt/internal/application"
	"github.com/gamee1910/volt/internal/application/command"
	"github.com/gamee1910/volt/internal/common/utils"
	"github.com/gamee1910/volt/pkg/logger"
	"github.com/robfig/cron/v3"
)

const (
	every12HoursCron = "0 */12 * * *"
)

type ElectricityScheduler struct {
	cron *cron.Cron
	app  *application.Application
	log  *logger.Logger
}

func NewElectricityScheduler(
	app *application.Application,
	log *logger.Logger,
) (*ElectricityScheduler, error) {
	location, err := utils.LoadVietnamTimezone()
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}
	option := cron.WithLocation(location)
	c := cron.New(option)

	s := &ElectricityScheduler{
		cron: c,
		app:  app,
		log:  log,
	}

	_, err = c.AddFunc(every12HoursCron, func() {
		if err := s.sync(context.Background()); err != nil {
			log.Error("electricity_sync_job: failed", "error", err.Error())
			return
		}
		log.Info("electricity_sync_job: success")
	})
	if err != nil {
		return nil, fmt.Errorf("register electricity scheduler cron error: %w", err)
	}

	return s, nil
}

func (s *ElectricityScheduler) sync(ctx context.Context) error {
	location, err := utils.LoadVietnamTimezone()
	if err != nil {
		return err
	}

	now := time.Now().In(location)
	firstDayOfMonth := utils.GetFirstDayOfMonth(now, location)

	return s.app.Commands.SyncElectricity.Handle(ctx, command.SyncElectricityCommand{
		FromDate: utils.FormatEVNDate(firstDayOfMonth),
		ToDate:   utils.FormatEVNDate(now),
	})
}

func (s *ElectricityScheduler) Start() {
	s.log.Infof("electricity_scheduler: started, schedule=%s", every12HoursCron)
	go func() {
		if err := s.sync(context.Background()); err != nil {
			s.log.Error("electricity_scheduler: initial run failed", "error", err.Error())
		}
	}()
	s.cron.Start()
}

func (s *ElectricityScheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.log.Info("electricity_scheduler: stopped")
}
