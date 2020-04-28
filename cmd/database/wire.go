//+build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/polyse/database/internal/web"
)

func initWebApp(c *config) (web.App, error) {
	wire.Build(web.NewApp, NewWebAppCfg)
	return web.App{}, nil
}
