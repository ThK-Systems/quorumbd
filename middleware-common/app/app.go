// Package app package is the main app for all middlewares
package app

import (
	logging "thk-systems.net/quorumbd/common/logging"
	config "thk-systems.net/quorumbd/middleware-common/config"
)

var (
	Logger = logging.For("middlewareapp")
	Config config.Config
)

type app struct {
	implementation string
}

func New(adaptor Adaptor, config config.Config) app {
	Logger = Logger.With("impl", adaptor.GetName())
	return app{
		implementation: adaptor.GetName(),
	}
}

func (app app) Run() error {
	Logger.Info("Middleware is about to start ...")
	Logger.Info("Middleware is about to exit ...")
	return nil
}
