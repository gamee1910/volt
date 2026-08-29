package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/gamee1910/volt/config"
	"github.com/gamee1910/volt/internal/application/command"
	"github.com/gamee1910/volt/internal/common/utils"
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

	location, err := utils.LoadVietnamTimezone()
	if err != nil {
		log.Fatalf("[sync] timezone error: ", "error", err)
	}

	now := time.Now().In(location)
	firstDayOfMonth := utils.GetFirstDayOfMonth(now, location)

	err = container.App().Commands.SyncElectricity.Handle(context.Background(), command.SyncElectricityCommand{
		FromDate: utils.FormatEVNDate(firstDayOfMonth),
		ToDate:   utils.FormatEVNDate(now),
	})
	if err != nil {
		log.Error("[sync] failed to sync electricity", "error", err)
	}
}

func closeDB(db *sql.DB, log *logger.Logger) {
	if err := db.Close(); err != nil {
		log.Error("[sync] failed to close database", "error", err)
	}
}
