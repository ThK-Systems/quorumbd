// Package app is the main app for all middlewares
package app

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"thk-systems.net/quorumbd/common/helper/synchelper"
	"thk-systems.net/quorumbd/middleware-common/config"
	"thk-systems.net/quorumbd/middleware-common/coreconnection"
)

var (
	Config config.Config
)

type App struct {
	logger      *slog.Logger
	config      *config.Config
	coreManager *coreconnection.CoreManager
	adaptor     Adaptor
}

func New(adaptor Adaptor, config *config.Config, logger *slog.Logger) (*App, error) {
	if logger == nil {
		logger = slog.Default()
	}
	ccm, err := coreconnection.New(&config.CoreConnectionConfig)
	if err != nil {
		return nil, err
	}
	newApp := App{
		logger:      logger,
		config:      config,
		coreManager: ccm,
		adaptor:     adaptor,
	}
	newApp.logger = newApp.logger.With("impl", newApp.adaptor.GetImplementationName())
	return &newApp, nil
}

func (app *App) Run() error {
	app.logger.Info("Middleware is about to start ...")
	tg := synchelper.TaskGroup{}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	errCh := make(chan error, 1)

	tg.Go(func() {
		mainLoop(ctx, errCh, app)
	})

	select {

	case err := <-errCh:
		return err

	case <-ctx.Done():
		tg.Wait() // Waiting for task group to complete

		app.logger.Info("Middleware is exiting ...")
		return nil

	}
}

func mainLoop(ctx context.Context, errCh chan<- error, app *App) {
	<-ctx.Done()
	errCh <- nil
}
