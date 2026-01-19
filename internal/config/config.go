package config

import (
	"fmt"
	"time"
)

type Config struct {
	HTTPPort string `env:"HTTP_PORT" default:"8080"`
	DBURL    string `env:"DB_URL" default:""`

	PostgresHost     string `env:"POSTGRES_HOST" default:"localhost"`
	PostgresPort     string `env:"POSTGRES_PORT" default:"5432"`
	PostgresUser     string `env:"POSTGRES_USER" default:"postgres"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" default:"postgres"`
	PostgresDatabase string `env:"POSTGRES_DATABASE" default:"subscriptions"`

	MaxOpenConns int           `env:"POSTGRES_MAX_OPEN_CONNECTIONS" default:"25"`
	MaxIdleConns int           `env:"POSTGRES_MAX_IDLE_CONNECTIONS" default:"25"`
	ConnMaxLife  time.Duration `env:"POSTGRES_CONNECTION_MAX_LIFETIME" default:"5m"`
}

func (c *Config) BuildDBURL() string {
	if c.DBURL != "" {
		return c.DBURL
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDatabase,
	)
}
