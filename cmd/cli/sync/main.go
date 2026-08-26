package main

import (
	"context"
	"database/sql"

	"github.com/gamee1910/volt/config"
	"github.com/gamee1910/volt/internal/infrastructure/di"
	"github.com/gamee1910/volt/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.NewLogger(cfg.ApplicationConfig.Env)

	databaseConnection, err := config.DatabaseConnection(cfg)
	if err != nil {
		log.Fatalf("[sync] database error: ", "error", err)
	}
	defer closeDB(databaseConnection, log)

	container, err := di.NewContainer(cfg, databaseConnection, log)
	if err != nil {
		log.Fatalf("[sync] container error: ", "error", err)
	}

	if err := container.ElectricityWorker().Run(context.Background()); err != nil {
		log.Error("[sync] failed to run electricity worker", "error", err)
	}

}
func closeDB(db *sql.DB, log *logger.Logger) {
	if err := db.Close(); err != nil {
		log.Error("[sync] failed to close database", "error", err)
	}
}
