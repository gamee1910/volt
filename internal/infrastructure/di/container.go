package di

import (
	"database/sql"
	"fmt"

	"github.com/gamee1910/volt/config"
	service2 "github.com/gamee1910/volt/internal/application"
	"github.com/gamee1910/volt/internal/domain/ports"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/internal/infrastructure/client"
	"github.com/gamee1910/volt/internal/infrastructure/persistences/postgres"
	"github.com/gamee1910/volt/internal/interfaces/api/handler"
)

type Container struct {
	cfg                *config.Configuration
	db                 *sql.DB
	evnClient          *client.EVNClient
	electricityHandler *handler.ElectricityHandler
}

func (c *Container) ElectricityHandler() *handler.ElectricityHandler {
	return c.electricityHandler
}

func NewContainer(cfg *config.Configuration, db *sql.DB) (*Container, error) {
	evnClient, err := client.NewEVNClient(
		cfg.ApplicationConfig.EnvConfig.BaseURL,
		cfg.ApplicationConfig.EnvConfig.LoginAPI,
		cfg.ApplicationConfig.EnvConfig.ElectricityConsumptionAPI,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create EVN client: %w", err)
	}

	c := &Container{
		cfg:       cfg,
		db:        db,
		evnClient: evnClient,
	}
	c.initializerHandler()
	return c, nil
}

func (c *Container) initializerHandler() {
	repositories := c.initRepositories()
	services := c.initServices(repositories)

	c.electricityHandler = handler.NewElectricityHandler(services.electricityService, c.cfg)
}

type repositories struct {
	electricityRepository ports.ElectricityRepository
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
		electricityService: service2.NewElectricityService(r.electricityRepository, c.evnClient),
	}
}
