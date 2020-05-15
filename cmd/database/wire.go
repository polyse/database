//+build wireinject

package main

import (
	"context"
	"github.com/google/wire"
	"github.com/polyse/database/internal/api"
)

func initWebApp(c *config) (*api.API, func(), error) {
	wire.Build(api.NewApp, initWebAppCfg)
	return nil, nil, nil
}
