package di

import (
	"database/sql"
	"fmt"

	"github.com/gamee1910/volt/internal/application"
	"github.com/gamee1910/volt/internal/config"
	"github.com/gamee1910/volt/internal/domain/repository"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/internal/infrastructure/client/evnhcm"
	"github.com/gamee1910/volt/internal/infrastructure/persistences/postgres"
	"github.com/gamee1910/volt/internal/interfaces/http/handler"
)

type Container struct {
	cfg                *config.Configuration
	db                 *sql.DB
	evnClient          *evnhcm.EVNClient
	electricityHandler *handler.ElectricityHandler
}

func (c *Container) ElectricityHandler() *handler.ElectricityHandler {
	return c.electricityHandler
}

func NewContainer(cfg *config.Configuration, db *sql.DB) (*Container, error) {
	evnClient, err := evnhcm.NewEVNClient(nil)
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
