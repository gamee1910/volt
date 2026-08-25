package di

import (
	"database/sql"
	"fmt"

	"github.com/gamee1910/volt/config"
	"github.com/gamee1910/volt/internal/application"
	"github.com/gamee1910/volt/internal/domain/repository"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/internal/infrastructure/persistences/postgres"
	"github.com/gamee1910/volt/internal/infrastructure/scheduler"
	"github.com/gamee1910/volt/internal/interfaces/api/handler"
	"github.com/gamee1910/volt/internal/interfaces/bot/command"
	"github.com/gamee1910/volt/internal/interfaces/bot/routes"
	"github.com/gamee1910/volt/internal/interfaces/worker"
	"github.com/gamee1910/volt/pkg/evnhcmc"
	"github.com/gamee1910/volt/pkg/logger"
	"github.com/gamee1910/volt/pkg/telegram"
)

type Container struct {
	configuration *config.Configuration
	database      *sql.DB
	logger        *logger.Logger

	// Clients
	telegramClient telegram.TelegramClient
	evnClient      evnhcmc.EVNClient

	// Handler
	electricityHandler *handler.ElectricityHandler

	// Command
	electricityCommand *command.ElectricityCommand

	// Bot
	botRouter *routes.Router

	//Worker
	electricityWorker *worker.ElectricityWorker

	//Scheduler
	electricityScheduler *scheduler.ElectricityScheduler
}

func (c *Container) TelegramClient() telegram.TelegramClient {
	return c.telegramClient
}

func (c *Container) EVNClient() evnhcmc.EVNClient {
	return c.evnClient
}

func (c *Container) ElectricityHandler() *handler.ElectricityHandler {
	return c.electricityHandler
}

func (c *Container) BotRouter() *routes.Router {
	return c.botRouter
}

func (c *Container) ElectricityCommand() *command.ElectricityCommand {
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
	client, err := initClients(cfg, log)
	if err != nil {
		return nil, err
	}

	container := &Container{
		configuration: cfg,
		database:      db,
		logger:        log,

		telegramClient: client.telegramClient,
		evnClient:      client.evnClient,
	}

	container.initializerHandler()
	return container, nil
}

func (c *Container) initializerHandler() {
	repositories := c.initRepositories()
	services := c.initServices(repositories)

	c.electricityHandler = handler.NewElectricityHandler(
		services.electricityService,
		c.configuration,
	)

	c.electricityWorker = worker.NewElectricityWorker(
		services.electricityService,
		c.logger,
	)

	scheduler, err := scheduler.NewElectricityScheduler(
		c.electricityWorker,
		c.logger,
	)
	if err != nil {
		c.logger.Fatal("failed to create scheduler", "error", err)
	}

	c.electricityScheduler = scheduler

	c.initializeCommand(services)
}

func (c *Container) initializeCommand(services services) {
	c.electricityCommand = command.NewElectricityCommand(
		c.configuration,
		c.logger,
		c.telegramClient,
		services.electricityService,
	)

	c.botRouter = routes.NewRouter(
		c.logger,
		c.telegramClient,
		c.electricityCommand,
	)

	c.telegramClient.SetMessageHandler(
		c.botRouter.DefaultHandler(),
	)
}

type repositories struct {
	electricityRepository repository.ElectricityRepository
}

func (c *Container) initRepositories() repositories {
	return repositories{
		electricityRepository: postgres.NewElectricityRepository(c.database),
	}
}

type services struct {
	electricityService service.ElectricityService
}

func (c *Container) initServices(r repositories) services {
	return services{
		electricityService: application.NewElectricityService(r.electricityRepository, c.evnClient, c.logger),
	}
}

type clients struct {
	telegramClient telegram.TelegramClient
	evnClient      evnhcmc.EVNClient
}

func initClients(cfg *config.Configuration, log *logger.Logger) (clients, error) {
	telegramClient, err := telegram.NewTelegramClient(
		cfg.ApplicationConfig.TelegramConfig.TelegramAPIKey,
		log,
	)
	if err != nil {
		return clients{}, fmt.Errorf(
			"failed to initialize telegram client: %w",
			err,
		)
	}

	evnClient, err := evnhcmc.NewEVNClient(cfg, log)
	if err != nil {
		return clients{}, fmt.Errorf(
			"failed to initialize evn client: %w",
			err,
		)
	}

	return clients{
		telegramClient: telegramClient,
		evnClient:      evnClient,
	}, nil
}
