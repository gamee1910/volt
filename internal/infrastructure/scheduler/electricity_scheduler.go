package scheduler

import (
	"context"
	"fmt"

	"github.com/gamee1910/volt/internal/common/utils"
	"github.com/gamee1910/volt/internal/interfaces/worker"
	"github.com/gamee1910/volt/pkg/logger"
	"github.com/robfig/cron/v3"
)

const (
	every12HoursCron = "0 */12 * * *"
)

type ElectricityScheduler struct {
	cron   *cron.Cron
	worker *worker.ElectricityWorker
	log    *logger.Logger
}

func NewElectricityScheduler(
	worker *worker.ElectricityWorker,
	log *logger.Logger,
) (
	*ElectricityScheduler,
	error,
) {
	location, err := utils.LoadVietnamTimezone()
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}
	option := cron.WithLocation(location)
	c := cron.New(option)

	_, err = c.AddFunc(every12HoursCron, func() {
		if err := worker.Run(context.Background()); err != nil {
			log.Error("eletricity_sync_job: failed", "error", err.Error())
			return
		}
		log.Info("eletricity_sync_job: success")
	})
	if err != nil {
		return nil, fmt.Errorf("register electricity scheduler cron error: %w", err)
	}

	return &ElectricityScheduler{
		cron:   c,
		worker: worker,
		log:    log,
	}, nil
}

func (s *ElectricityScheduler) Start() {
	s.log.Infof("eletricity_scheduler: started, schedule=%s\n", every12HoursCron)
	go func() {
		if err := s.worker.Run(context.Background()); err != nil {
			s.log.Error("eletricity_scheduler: initial run failed", "error", err.Error())
		}
	}()
	s.cron.Start()
}

func (s *ElectricityScheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.log.Info("eletricity_scheduler: stopped")
}
