// Package app is the main app for all middlewares
package app

import (
	logging "thk-systems.net/quorumbd/common/logging"
	config "thk-systems.net/quorumbd/middleware-common/config"
)

var (
	Logger = logging.For("middlewareapp")
	Config config.Config
)

type App struct {
	Config  config.Config
	Adaptor Adaptor
}

func New(adaptor Adaptor, config config.Config) App {
	newApp := App{
		Config:  config,
		Adaptor: adaptor,
	}
	Logger = Logger.With("impl", newApp.Adaptor.GetImplementationName())
	return newApp
}

func (app App) Run() error {
	Logger.Info("Middleware is about to start ...")
	Logger.Info("Middleware is exiting ...")
	return nil
}
