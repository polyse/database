//+build wireinject

package main

import (
	"context"

	"github.com/polyse/database/internal/collection"

	"github.com/google/wire"
	"github.com/polyse/database/internal/api"
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

func initWebApp(ctx context.Context, c *config) (*api.API, func(), error) {
	wire.Build(api.NewApp, initWebAppCfg)
	return nil, nil, nil
}

func initProcessorManager(
	c *config,
	collName collection.Name,
) (*collection.Manager, func(), error) {
	wire.Build(dbSetter)
	return nil, nil, nil
}
