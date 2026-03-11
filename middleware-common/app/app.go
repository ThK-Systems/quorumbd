// Package app is the main app for all middlewares
package app

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"quorumbd.net/common/helper/synchelper"
	"quorumbd.net/middleware-common/config"
	"quorumbd.net/middleware-common/coreconnection"
)

var (
	Config config.Config
)

type App struct {
	logger         *slog.Logger
	config         *config.Config
	coreSupervisor *coreconnection.CoreSupervisor
	adaptor        Adaptor
}

func New(adaptor Adaptor, config *config.Config, logger *slog.Logger) (*App, error) {
	if logger == nil {
		logger = slog.Default()
	}

	cs, err := coreconnection.New(&config.CoreConnectionConfig, logger)
	if err != nil {
		return nil, err
	}

	newApp := App{
		logger:         logger,
		config:         config,
		coreSupervisor: cs,
		adaptor:        adaptor,
	}

	newApp.logger = newApp.logger.With("impl", newApp.adaptor.GetImplementationName())
	return &newApp, nil
}

func (app *App) Run() error {
	app.logger.Info("Middleware is about to start ...")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	errCh := make(chan error, 1)

	tg := synchelper.TaskGroup{}
	tg.Go(func() {
		errCh <- mainLoop(ctx, app)
	})

	var err error

	select {
	case err = <-errCh:
		stop()
	case <-ctx.Done():
		stop()
		err = <-errCh
	}

	tg.Wait()

	if err == nil || errors.Is(err, context.Canceled) {
		app.logger.Info("Middleware is exiting ...")
		return nil
	}
	return err
}

func mainLoop(ctx context.Context, app *App) error {
	if _, err := app.coreSupervisor.Probe(ctx, 0, 30*time.Second, false, false); err != nil { // DO NOT USE Probe, use Connect or something like that and make Probe private
		return err
	}

	// TODO: Start ControlSuperVisor
	// TODO: Start DiskSuperVisor

	<-ctx.Done()
	return ctx.Err()
}
