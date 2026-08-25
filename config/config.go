package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type (
	// Config -.
	Config struct {
		App  app
		HTTP http
		Log  log
		PG   pg
	}

	// App -.
	app struct {
		Name    string `env:"APP_NAME" envDefault:"mockzoo"`
		Version string `env:"APP_VERSION" envDefault:"1.0.0"`
	}

	// HTTP -.
	http struct {
		Port           string `env:"HTTP_PORT" envDefault:"8080"`
		UsePreforkMode bool   `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
	}

	// Log -.
	log struct {
		Level string `env:"LOG_LEVEL" envDefault:"info"`
	}

	// PG -.
	pg struct {
		PoolMax int    `env:"PG_POOL_MAX" envDefault:"2"`
		URL     string `env:"PG_URL,required"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
