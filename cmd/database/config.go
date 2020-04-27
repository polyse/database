package main

import (
	"github.com/caarlos0/env"
)

type config struct {
	Listen   string `env:"LISTEN" envDefault:"localhost:9000"`
	Timeout  string `env:"TIMEOUT" envDefault:"10ms"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
	LogFmt   string `env:"LOG_FMT" envDefault:"console"`
}

func Load() (*config, error) {
	cfg := &config{}

	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
