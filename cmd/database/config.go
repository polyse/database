package main

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/caarlos0/env"
)

// Config is main application configuration structure.
type config struct {
	Listen       string        `env:"LISTEN" envDefault:"localhost:9000"`
	Timeout      time.Duration `env:"TIMEOUT" envDefault:"10ms"`
	MergeTimeout time.Duration `env:"TIMEOUT" envDefault:"1h"`
	LogLevel     string        `env:"LOG_LEVEL" envDefault:"info"`
	LogFmt       string        `env:"LOG_FMT" envDefault:"console"`
	DbFile       string        `env:"DB_FILE" envDefault:"./tmp/nutsdb"`
}

func load() (*config, error) {
	log.Debug().Msg("loading configuration")
	cfg := &config{}

	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
