//+build wireinject

package main

import (
	"context"
	"github.com/google/wire"
	"github.com/polyse/database/internal/web"
)

func initWebApp(c *config) (*web.API, func(), error) {
	wire.Build(web.NewApp, initWebAppCfg)
	return nil, nil, nil
}
