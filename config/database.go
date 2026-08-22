package config

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

func DatabaseConnection(cfg *Configuration) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseConfig.DatabaseDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DatabaseConfig.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.DatabaseConfig.MaxIdleConnections)
	db.SetConnMaxIdleTime(cfg.DatabaseConfig.MaxIdleTime)

	ctx, cancle := context.WithTimeout(context.Background(), cfg.ServerConfig.ReadTimeout)
	defer cancle()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
