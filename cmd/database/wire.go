//+build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/polyse/database/internal/proc"
	"github.com/polyse/database/internal/web"
)

func initWebApp(ctx context.Context, c *config) (*web.App, func(), error) {
	wire.Build(web.NewApp, initWebAppCfg)
	return nil, nil, nil
}

func initProcessorManager(c *config) (*proc.SimpleProcessorManager, func(), error) {
	wire.Build(dbSetter)
	return nil, nil, nil
}
