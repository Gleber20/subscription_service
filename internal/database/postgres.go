package database

import (
	"fmt"

	"subscription_service/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgres(cfg *config.Config) (*sqlx.DB, error) {
	dsn := cfg.BuildDBURL()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlx open error: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLife)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("db ping error: %w", err)
	}

	return db, nil
}
