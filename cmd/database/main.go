package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/wire"
	"github.com/polyse/database/internal/db"
	"github.com/polyse/database/internal/proc"

	"github.com/polyse/database/internal/web"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xlab/closer"
)

var (
	procSetter = wire.NewSet(
		initDbCol,
		initDbCfg,
		db.NewConnection,
		db.NewNutRepo,
		wire.Bind(
			new(db.Repository),
			new(*db.NutsRepository)),
		proc.NewProcessor,
	)

	dbSetter = wire.NewSet(
		procSetter,
		wire.Bind(
			new(proc.Processor),
			new(*proc.SimpleProcessor),
		),
		proc.NewSimpleProcessorManagerWithProc,
	)
)

func main() {

	defer closer.Close()

	ctx, cancelCtx := context.WithCancel(context.Background())

	closer.Bind(cancelCtx)

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
	a, cancel, err := initWebApp(ctx, cfg)
	if err != nil {
		log.Err(err).Msg("error while init wire")
		return
	}

	log.Debug().Msg("starting web application")

	// Bind closer func to smoothly close connection.
	closer.Bind(cancel)

	_, dbCancel, err := initProcessorManager(cfg)

	closer.Bind(dbCancel)

	log.Debug().Msg("starting web application")

	if err = a.Run(); err != nil {
		log.Err(err).Msg("error while starting web app")
	}
}

func initDbCol(c *config) db.CollectionName {
	return db.CollectionName(c.BaseCollection)
}

func initDbCfg(c *config) db.Config {
	return db.Config(c.DbFile)
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
