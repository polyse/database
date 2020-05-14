package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/polyse/database/internal/web"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xlab/closer"
)

func main() {

	defer closer.Close()

	cfg, err := load()

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
	a, cancel, err := initWebApp(cfg)
	if err != nil {
		log.Err(err).Msg("error while init wire")
		return
	}

	log.Debug().Msg("starting web application")

	// Bind closer func to smoothly close connection.
	closer.Bind(cancel)

	if err = a.Run(); err != nil {
		log.Err(err).Msg("error while starting web app")
	}
}

func initWebAppCfg(c *config) (web.AppConfig, error) {
	return web.AppConfig{Timeout: c.Timeout, NetInterface: c.Listen}, nil
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
