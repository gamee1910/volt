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

// SyncNotifier is told the result of each scheduled sync.
type SyncNotifier interface {
	NotifySyncResult(ctx context.Context, syncErr error) error
}

type ElectricityScheduler struct {
	cron     *cron.Cron
	app      *application.Application
	log      *logger.Logger
	notifier SyncNotifier
}

func NewElectricityScheduler(
	app *application.Application,
	log *logger.Logger,
	notifier SyncNotifier,
) (*ElectricityScheduler, error) {
	location, err := utils.LoadVietnamTimezone()
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}
	option := cron.WithLocation(location)
	c := cron.New(option)

	s := &ElectricityScheduler{
		cron:     c,
		app:      app,
		log:      log,
		notifier: notifier,
	}

	_, err = c.AddFunc(every12HoursCron, func() {
		ctx := context.Background()
		syncErr := s.sync(ctx)
		if syncErr != nil {
			log.Error("electricity_sync_job: failed", "error", syncErr.Error())
		} else {
			log.Info("electricity_sync_job: success")
		}
		s.notify(ctx, syncErr)
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

func (s *ElectricityScheduler) notify(ctx context.Context, syncErr error) {
	if s.notifier == nil {
		return
	}
	if err := s.notifier.NotifySyncResult(ctx, syncErr); err != nil {
		s.log.Error("electricity_sync_job: notify failed", "error", err.Error())
	}
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
