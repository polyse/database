package main

import (
	"fmt"
	"github.com/polyse/database/internal/db"
	"github.com/polyse/database/internal/web"
	"github.com/polyse/database/pkg/filters"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

func initTokenizer() filters.Tokenizer {
	log.Debug().Msg("initialize tokenizer")
	return filters.FilterText
}

func initFilters() []filters.Filter {
	log.Debug().Msg("initialize filters")
	return []filters.Filter{filters.StemmAndToLower, filters.StopWords}
}

func initDbCollection(c *config) db.CollectionName {
	return db.CollectionName(c.BaseCollection)
}

func initDbConfig(c *config) db.Config {
	return db.Config(c.DbFile)
}

func initWebAppCfg(c *config) (web.AppConfig, error) {
	return web.AppConfig{Timeout: c.Timeout, NetInterface: c.Listen}, nil
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
