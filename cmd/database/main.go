package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/polyse/database/internal/api"
	"github.com/polyse/database/pkg/filters"
	"github.com/rs/zerolog"
	"github.com/xujiajun/nutsdb"

	"github.com/polyse/database/internal/collection"

	"github.com/rs/zerolog/log"
	"github.com/xlab/closer"
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

	// Bind closer func to smoothly close connection.
	closer.Bind(cancel)

	log.Debug().Msg("starting db")
	var connCLoser func()
	a.Manager, connCLoser, err = initProcessorManager(cfg, "default")
	if err != nil {
		log.Err(err).Msg("can not init proc manager")
		return
	}
	closer.Bind(connCLoser)

	log.Debug().Msg("starting web application")
	if err = a.Run(); err != nil {
		log.Err(err).Msg("error while starting api app")
	}
}

func initTokenizer() filters.Tokenizer {
	log.Debug().Msg("initialize tokenizer")
	return filters.FilterText
}

func initFilters() []filters.Filter {
	log.Debug().Msg("initialize filters")
	return []filters.Filter{filters.StemmAndToLower, filters.StopWords}
}

func initConnection(cfg collection.Config) (*nutsdb.DB, func(), error) {
	log.Debug().Interface("configuration", cfg).Msg("opening new connection to database")

	opt := nutsdb.DefaultOptions
	opt.Dir = cfg.File
	nutsDb, err := nutsdb.Open(opt)
	if err != nil {
		return nil, nil, err
	}
	log.Info().Msg("connection opened")
	return nutsDb, func() {
		log.Info().Msg("start closing database connection")
		if err = nutsDb.Merge(); err != nil {
			log.Err(err).Msg("can not merge database")
		}
		if err = nutsDb.Close(); err != nil {
			log.Err(err).Msg("can not close database connection")
		}
	}, nil
}

func initDbConfig(c *config) collection.Config {
	return collection.Config{File: c.DbFile}
}

func initWebAppCfg(c *config) (api.AppConfig, error) {
	return api.AppConfig{Timeout: c.Timeout, NetInterface: c.Listen}, nil
}

func initLogger(c *config) error {
	log.Debug().Msg("initialize logger")
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
