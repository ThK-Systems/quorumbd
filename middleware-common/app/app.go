// Package app is the main app for all middlewares
package app

import (
	"log/slog"

	config "thk-systems.net/quorumbd/middleware-common/config"
)

var (
	Config config.Config
)

type App struct {
	Logger  *slog.Logger
	Config  *config.Config
	Adaptor Adaptor
}

func New(adaptor Adaptor, config *config.Config, logger *slog.Logger) App {
	if logger == nil {
		logger = slog.Default()
	}
	newApp := App{
		Logger:  logger,
		Config:  config,
		Adaptor: adaptor,
	}
	newApp.Logger = newApp.Logger.With("impl", newApp.Adaptor.GetImplementationName())
	return newApp
}

func (app App) Run() error {
	app.Logger.Info("Middleware is about to start ...")
	// TODO - Implement middleware application
	app.Logger.Info("Middleware is exiting ...")
	return nil
}
