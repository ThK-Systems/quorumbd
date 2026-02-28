// Package middlewareapp package is the main app for all middlewares
package middlewareapp

import (
	logging "thk-systems.net/quorumbd/common/logging"
)

var (
	logger = logging.For("middlewareapp")
)

type app struct {
	implementation string
}

func New(implvalue string) app {
	logger = logger.With("impl", implvalue)
	return app{
		implementation: implvalue,
	}
}

func (app app) Run() error {
	logger.Info("Middleware is about to start ...")
	logger.Info("Middleware is about to exit ...")
	return nil
}
