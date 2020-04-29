//+build wireinject

package main

import (
	"context"
	"github.com/google/wire"
	"github.com/polyse/database/internal/web"
)

func initWebApp(ctx context.Context, c *config) (*web.App, func(), error) {
	wire.Build(web.NewApp, NewWebAppCfg)
	return nil, nil, nil
}
