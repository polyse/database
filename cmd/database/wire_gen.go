// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"github.com/polyse/database/internal/api"
)

// Injectors from wire.go:

func initWebApp(c *config) (*api.API, func(), error) {
	appConfig, err := initWebAppCfg(c)
	if err != nil {
		return nil, nil, err
	}
	app, err := api.NewApp(appConfig)
	if err != nil {
		return nil, nil, err
	}
	return app, func() {
		app.Close()
	}, nil
}
