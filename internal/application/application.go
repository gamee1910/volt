package application

import (
	"github.com/gamee1910/volt/internal/application/command"
	"github.com/gamee1910/volt/internal/application/query"
	"github.com/gamee1910/volt/internal/common/decorator"
	"github.com/gamee1910/volt/internal/domain/repository"
	"github.com/gamee1910/volt/pkg/evnhcmc"
	"github.com/gamee1910/volt/pkg/logger"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	SyncElectricity command.SyncElectricityHandler
}

type Queries struct {
	AllElectricity query.AllElectricityHandler
	YesterdayUsage query.YesterdayHandler
}

func NewApplication(
	evnClient evnhcmc.EVNClient,
	electricityRepository repository.ElectricityRepository,
	logger *logger.Logger,
	metricClient decorator.MetricClient,
) Application {
	return Application{
		Commands: Commands{
			SyncElectricity: command.NewSyncElectricityHandler(electricityRepository, evnClient, logger, metricClient),
		},
		Queries: Queries{
			AllElectricity: query.NewAllElectricityHandler(electricityRepository, logger, metricClient),
			YesterdayUsage: query.NewYesterdayUsageHandler(electricityRepository, logger, metricClient),
		},
	}
}
