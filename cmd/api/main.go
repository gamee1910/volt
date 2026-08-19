package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gamee1910/volt/internal/config"
	"github.com/gamee1910/volt/internal/routes"
	"github.com/gamee1910/volt/internal/routes/client/evnhcmc"
	"github.com/gamee1910/volt/pkg/logger"
)

func main() {

	cfg := config.Load()
	log := logger.Init(cfg.ApplicationConfig.Env)

	databaseConnection, err := config.DatabaseConnection(cfg)
	if err != nil {
		log.Fatal("database error ", "error", err)
	}
	defer closeDB(databaseConnection, log)

	evnClient, err := evnhcmc.NewEVNClient()
	if err != nil {
		logger.Fatal("Không thể khởi tạo EVN client", "error", err)
	}
	router := routes.SetupRouter(databaseConnection, evnClient)

	server := &http.Server{
		Addr:         ":" + cfg.ServerConfig.Port,
		Handler:      router,
		ReadTimeout:  cfg.ServerConfig.ReadTimeout,
		WriteTimeout: cfg.ServerConfig.WriteTimeout,
		IdleTimeout:  time.Minute,
	}

	go func() {
		log.Infof(
			"starting application [%s] port [%s] env [%s]",
			cfg.ApplicationConfig.Name,
			cfg.ServerConfig.Port,
			cfg.ApplicationConfig.Env,
		)

		var err error

		if cfg.ServerConfig.TLS.Mode == "enabled" {
			err = server.ListenAndServeTLS(
				cfg.ServerConfig.TLS.CertFile,
				cfg.ServerConfig.TLS.KeyFile,
			)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	gracefulShutdown(server, log)
}
func closeDB(db *sql.DB, log *logger.Logger) {
	if err := db.Close(); err != nil {
		log.Error("failed to close database", "error", err)
	}
}

func gracefulShutdown(server *http.Server, log *logger.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Info("server stopped")
}
