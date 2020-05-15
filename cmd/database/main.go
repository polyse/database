package main

import (
	"context"

	"github.com/google/wire"
	"github.com/polyse/database/internal/collection"

	"github.com/rs/zerolog/log"
	"github.com/xlab/closer"
)

var (
	procSetter = wire.NewSet(
		initDbConfig,
		initConnection,
		initTokenizer,
		initFilters,
		collection.NewSimpleProcessor,
	)

	dbSetter = wire.NewSet(
		procSetter,
		wire.Bind(
			new(collection.Processor),
			new(*collection.SimpleProcessor),
		),
		collection.NewManagerWithProc,
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

	log.Debug().Msg("starting db")
	_, connCLoser, err := initProcessorManager(cfg, "default")
	closer.Bind(connCLoser)

	log.Debug().Msg("starting web application")

	if err = a.Run(); err != nil {
		log.Err(err).Msg("error while starting web app")
	}
}
