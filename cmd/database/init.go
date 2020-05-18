package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/polyse/database/internal/collection"
	"github.com/polyse/database/internal/api"
	"github.com/polyse/database/pkg/filters"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xujiajun/nutsdb"
)

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
