package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/polyse/database/internal/web"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xlab/closer"
)

func main() {

	defer closer.Close()

	cfg, err := Load()
	if err != nil {
		log.Err(err).Msg("error while loading config")
		return
	}
	if err = initLogger(cfg); err != nil {
		log.Err(err).Msg("error while configure logger")
		return
	}
	log.Info().Msg(` 
	 ____       _       ____  _____   ____  ____  
	|  _ \ ___ | |_   _/ ___|| ____| |  _ \| __ ) 
	| |_) / _ \| | | | \___ \|  _|   | | | |  _ \ 
	|  __/ (_) | | |_| |___) | |___  | |_| | |_) |
	|_|   \___/|_|\__, |____/|_____| |____/|____/ 
				  |___/                         
	`)
	log.Debug().Msg("logger initialized")

	log.Debug().Msg("starting di container")
	a, err := InitializeEvent(cfg)
	if err != nil {
		log.Err(err).Msg("error while init wire")
		return
	}

	log.Debug().Msg("starting web application")
	if err = a.Run(context.Background()); err != nil && err != http.ErrServerClosed {
		log.Err(err).Msg("error while starting web app")
	}
}

func NewWebAppCfg(c *config) (web.AppConfig, error) {
	tmt, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return web.AppConfig{}, err
	}
	return web.AppConfig{Timeout: tmt, NetInterface: c.Listen}, nil
}

func initLogger(c *config) error {
	logLvl, err := zerolog.ParseLevel(strings.ToLower(c.LogLevel))
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLvl)
	switch c.LogFmt {
	case "console":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	case "json":
	default:
		return fmt.Errorf("unknown output format %s", c.LogFmt)
	}
	return nil
}
