// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"context"
	"github.com/polyse/database/internal/db"
	"github.com/polyse/database/internal/proc"
	"github.com/polyse/database/internal/web"
)

// Injectors from wire.go:

func initWebApp(ctx context.Context, c *config) (*web.App, func(), error) {
	appConfig, err := initWebAppCfg(c)
	if err != nil {
		return nil, nil, err
	}
	app, cleanup, err := web.NewApp(ctx, appConfig)
	if err != nil {
		return nil, nil, err
	}
	return app, func() {
		cleanup()
	}, nil
}

func initProcessorManager(c *config) (proc.SimpleProcessorManager, func(), error) {
	collectionName := initDbCol(c)
	dbConfig := initDbCfg(c)
	connection, cleanup, err := db.NewConnection(dbConfig)
	if err != nil {
		return nil, nil, err
	}
	nutsRepository := db.NewNutRepo(collectionName, connection)
	simpleProcessor := proc.NewProcessor(nutsRepository)
	simpleProcessorManager := proc.NewSimpleProcessorManagerWithProc(simpleProcessor)
	return simpleProcessorManager, func() {
		cleanup()
	}, nil
}
