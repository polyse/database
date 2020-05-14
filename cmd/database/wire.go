//+build wireinject

package main

import (
	"context"
	"github.com/google/wire"
	"github.com/polyse/database/internal/web"
)

func initWebApp(c *config) (*web.App, func(), error) {
	wire.Build(web.NewApp, initWebAppCfg)
	return nil, nil, nil
}
