package di

import (
	"database/sql"
	"net/url"

	"github.com/gamee1910/volt/internal/application"
	"github.com/gamee1910/volt/internal/client"
	"github.com/gamee1910/volt/internal/config"
	"github.com/gamee1910/volt/internal/domain/repository"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/internal/infrastructure/client/evn"
	"github.com/gamee1910/volt/internal/infrastructure/persistences/postgres"
)

const (
	defaultBaseURL = ""
)

type Container struct {
	cfg *config.Configuration
	db  *sql.DB

	//handler

	//client
	evnClient client.EvnhcmcClient
}

func NewContainer(cfg *config.Configuration, db *sql.DB) (*Container, error) {
	c := &Container{cfg: cfg, db: db}

	if err := c.initEVNHCMCClient(); err != nil {
		return nil, err
	}

	c.initializerHandler()

	return c, nil
}

func (c *Container) EVNClient() client.EvnhcmcClient {
	return c.evnClient
}

func (c *Container) initializerHandler() {
	repositories := c.initRepositories()
	c.initServices(repositories)
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
		electricityService: application.NewElectricityService(r.electricityRepository),
	}
}

func (c *Container) initEVNHCMCClient() error {
	parsedURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return err
	}

	evnClient, err := evn.NewEVNClient(parsedURL)
	if err != nil {
		return err
	}

	c.evnClient = evnClient
	return nil
}
