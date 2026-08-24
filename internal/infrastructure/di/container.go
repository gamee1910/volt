package di

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gamee1910/volt/config"
	"github.com/gamee1910/volt/internal/application"
	"github.com/gamee1910/volt/internal/application/port"
	"github.com/gamee1910/volt/internal/domain/repository"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/internal/infrastructure/client"
	"github.com/gamee1910/volt/internal/infrastructure/persistences/postgres"
	"github.com/gamee1910/volt/internal/interfaces/api/handler"
	"github.com/gamee1910/volt/internal/interfaces/bot/command"
	"github.com/gamee1910/volt/internal/interfaces/bot/router"
	"github.com/gamee1910/volt/pkg/logger"
	"go.uber.org/zap"
)

type Container struct {
	cfg *config.Configuration
	db  *sql.DB
	log *logger.Logger
	//Client
	evnClient      *client.EVNClient
	telegramClient port.TelegramClient

	//Handler
	electricityHandler *handler.ElectricityHandler
}

func (c *Container) ElectricityHandler() *handler.ElectricityHandler {
	return c.electricityHandler
}

func (c *Container) TelegramClient() port.TelegramClient {
	return c.telegramClient
}

func (c *Container) HTTPServer() *http.Server {
	return c.HTTPServer()
}

func NewContainer(cfg *config.Configuration, db *sql.DB, log *logger.Logger) (*Container, error) {
	evnClient, err := client.NewEVNClient(
		cfg.ApplicationConfig.EnvConfig.BaseURL,
		cfg.ApplicationConfig.EnvConfig.LoginAPI,
		cfg.ApplicationConfig.EnvConfig.ElectricityConsumptionAPI,
	)
	if err != nil {
		log.Fatal("failed to create evn client", zap.Error(err))
		return nil, fmt.Errorf("failed to create EVN client: %w", err)
	}

	c := &Container{
		cfg:       cfg,
		db:        db,
		log:       log,
		evnClient: evnClient,
	}

	c.initializerHandler()
	return c, nil
}

func (c *Container) initializerHandler() {
	repositories := c.initRepositories()
	services := c.initServices(repositories)

	electricityCommand := command.NewElectricityCommand(c.cfg, c.log, nil, services.electricityService)

	botRouter := router.NewRouter(c.log, nil, electricityCommand)

	telegramClient, err := client.NewTelegramClient(
		c.cfg.ApplicationConfig.TelegramConfig.TelegramAPIKey,
		c.log,
		botRouter.DefaultHandler(),
	)
	if err != nil {
		c.log.Fatalf("failed to create Telegram client: %v", err)
	}

	electricityCommand.SetSender(telegramClient)
	botRouter.SetSender(telegramClient)

	c.telegramClient = telegramClient
	c.electricityHandler = handler.NewElectricityHandler(services.electricityService, c.cfg)
}

type repositories struct {
	electricityRepository repository.ElectricityRepository
}

func (c *Container) initRepositories() repositories {
	return repositories{
		electricityRepository: postgres.NewElectricityRepository(c.db),
	}
}

type services struct {
	electricityService service.ElectricityService
}

func (c *Container) initServices(r repositories) services {
	return services{
		electricityService: application.NewElectricityService(r.electricityRepository, c.evnClient),
	}
}
