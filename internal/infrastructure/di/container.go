package di

import (
	"database/sql"
	"fmt"

	"github.com/gamee1910/volt/config"
	"github.com/gamee1910/volt/internal/application"
	"github.com/gamee1910/volt/internal/common/metrics"
	"github.com/gamee1910/volt/internal/infrastructure/persistences/postgres"
	"github.com/gamee1910/volt/internal/infrastructure/scheduler"
	"github.com/gamee1910/volt/internal/interfaces/api/handler"
	botcommand "github.com/gamee1910/volt/internal/interfaces/bot/command"
	"github.com/gamee1910/volt/internal/interfaces/bot/routes"
	"github.com/gamee1910/volt/internal/interfaces/worker"
	"github.com/gamee1910/volt/pkg/evnhcmc"
	"github.com/gamee1910/volt/pkg/logger"
	"github.com/gamee1910/volt/pkg/telegram"
)

type Container struct {
	configuration *config.Configuration
	logger        *logger.Logger

	telegramClient telegram.TelegramClient
	evnClient      evnhcmc.EVNClient

	electricityHandler   *handler.ElectricityHandler
	electricityCommand   *botcommand.ElectricityCommand
	electricityWorker    *worker.ElectricityWorker
	electricityScheduler *scheduler.ElectricityScheduler
}

func (c *Container) TelegramClient() telegram.TelegramClient {
	return c.telegramClient
}

func (c *Container) ElectricityHandler() *handler.ElectricityHandler {
	return c.electricityHandler
}

func (c *Container) ElectricityCommand() *botcommand.ElectricityCommand {
	return c.electricityCommand
}

func (c *Container) ElectricityWorker() *worker.ElectricityWorker {
	return c.electricityWorker
}

func (c *Container) ElectricityScheduler() *scheduler.ElectricityScheduler {
	return c.electricityScheduler
}

func NewContainer(
	cfg *config.Configuration,
	db *sql.DB,
	log *logger.Logger,
) (*Container, error) {
	telegramClient, evnClient, err := initClients(cfg, log)
	if err != nil {
		return nil, err
	}

	c := &Container{
		configuration:  cfg,
		logger:         log,
		telegramClient: telegramClient,
		evnClient:      evnClient,
	}

	if err := c.wire(db); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Container) wire(db *sql.DB) error {
	repo := postgres.NewElectricityRepository(db)
	metricClient := metrics.NewStatsClient(c.logger)

	app := application.NewApplication(
		c.evnClient,
		repo,
		c.logger,
		metricClient,
	)

	c.electricityHandler = handler.NewElectricityHandler(&app, c.configuration)
	c.electricityWorker = worker.NewElectricityWorker(&app, c.logger)

	sched, err := scheduler.NewElectricityScheduler(c.electricityWorker, c.logger)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}
	c.electricityScheduler = sched

	c.electricityCommand = botcommand.NewElectricityCommand(
		c.configuration,
		c.logger,
		c.telegramClient,
		&app,
	)

	botRouter := routes.NewRouter(c.logger, c.telegramClient, c.electricityCommand)
	c.telegramClient.SetMessageHandler(botRouter.DefaultHandler())

	return nil
}

func initClients(
	cfg *config.Configuration,
	log *logger.Logger,
) (telegram.TelegramClient, evnhcmc.EVNClient, error) {
	telegramClient, err := telegram.NewTelegramClient(
		cfg.ApplicationConfig.TelegramConfig.TelegramAPIKey,
		log,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize telegram client: %w", err)
	}

	evnClient, err := evnhcmc.NewEVNClient(cfg, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize evn client: %w", err)
	}

	return telegramClient, evnClient, nil
}
