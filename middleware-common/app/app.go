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
	dispatcher     *dispatcher
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

	dispatcher := newDispatcher(&newApp)
	newApp.dispatcher = dispatcher

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

	if err := app.coreSupervisor.Try(ctx, 0, 30*time.Second, false); err != nil {
		return err
	}

	errCh := make(chan error, 1)

	tg := synchelper.TaskGroup{}
	tg.Go(func() {
		errCh <- app.dispatcher.run(ctx)
	})

	// TODO: Do Listen here
	// TODO: Start DiskSuperVisor as go routine

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

func controlLoop(ctx context.Context, app *App) error {

	// TODO: Start ControlSuperVisor

	<-ctx.Done()
	return ctx.Err()
}
