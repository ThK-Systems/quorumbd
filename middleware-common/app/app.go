// Package app is the main app for all middlewares
package app

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"quorumbd.net/common/helper/errorhelper"
	"quorumbd.net/middleware-common/config"
	"quorumbd.net/middleware-common/coreconnection"
)

var (
	Config config.Config
)

type App struct {
	uuid           uuid.UUID
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
		uuid:           uuid.New(),
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

	if err := app.coreSupervisor.Try(ctx, 0, 30*time.Second, false); err != nil { // TOCONFIG
		return err
	}

	if !app.coreSupervisor.IsConnected() { // Test for initial core connection (Only relevant if initial core connection process was interrupted / shut down)
		app.logger.Warn("Middleware is exiting because initial core connection process has been interrupted")
		return nil
	}

	var (
		g                 errgroup.Group
		workerExitChannel = make(chan WorkerExit, 32)
		runError          error
	)

	g.Go(func() error {
		return app.dispatcher.run(ctx, workerExitChannel)
	})

	// TODO:
	// - Ask for disklist from core
	// - Ask for nodes from core

	// TODO: Do Listen here in go routine and start disk worker associated to errgroup
	// Do only listen, if server, otherwise connect proactively
	// Do not listen or connect proactively, if there is no core connection

outer:
	for {
		select {
		case <-ctx.Done():
			break outer
		case workerExitResult := <-workerExitChannel:
			switch workerExitResult.kind {
			case errorhelper.ExitFatal:
				runError = workerExitResult.err
				stop()
				break outer
			case errorhelper.ExitShutdown:
				stop()
				break outer
			case errorhelper.ExitReconnect:
				reconnectToCore(ctx, workerExitResult, workerExitChannel)
			}
		}

	}

	err := g.Wait() // Wait for all go routines to finish

	if runError != nil {
		err = runError // Fatal error (first error) overrides other errors
	}

	if err == nil || errors.Is(err, context.Canceled) {
		app.logger.Info("Middleware is exiting ...")
		return nil
	}

	app.logger.Error("Middleware is exiting with error", "error", err)
	return err
}

func reconnectToCore(ctx context.Context, workerExitResult WorkerExit, workerExitChannel chan<- WorkerExit) {
	// TODO Implement
	panic("unimplemented")
}
