// Package app is the main app for all middlewares
package app

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"quorumbd.net/common/helper/errorhelper"
	"quorumbd.net/middleware-common/config"
	"quorumbd.net/middleware-common/control"
	"quorumbd.net/middleware-common/coreconnection"
	"quorumbd.net/middleware-common/worker"
)

var (
	Config       config.Config
	appSingleton struct {
		mu          sync.Mutex
		initialized bool
	}
)

type App struct {
	uuid           uuid.UUID
	logger         *slog.Logger
	config         *config.Config
	coreSupervisor *coreconnection.CoreSupervisor
	adaptor        Adaptor
	dispatcher     *control.Dispatcher
	controlWorker  *control.ControlWorker
}

func New(adaptor Adaptor, config *config.Config, logger *slog.Logger) (*App, error) {
	claimAppSingleton()

	if logger == nil {
		logger = slog.Default()
	}

	cs, err := coreconnection.New(&config.CoreConnectionConfig, logger)
	if err != nil {
		releaseAppSingleton()
		return nil, err
	}

	newApp := App{
		uuid:           uuid.New(),
		logger:         logger,
		config:         config,
		coreSupervisor: cs,
		adaptor:        adaptor,
	}

	dispatcher := control.NewDispatcher(logger)
	newApp.dispatcher = dispatcher

	controlWorker := control.NewControlWorker(logger, dispatcher)
	newApp.controlWorker = controlWorker

	newApp.logger = newApp.logger.With("impl", newApp.adaptor.GetImplementationName())
	return &newApp, nil
}

func claimAppSingleton() {
	appSingleton.mu.Lock()
	defer appSingleton.mu.Unlock()

	if appSingleton.initialized {
		panic("app singleton already instantiated")
	}

	appSingleton.initialized = true
}

func releaseAppSingleton() {
	appSingleton.mu.Lock()
	defer appSingleton.mu.Unlock()
	appSingleton.initialized = false
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
		errgrp            errgroup.Group
		workerExitChannel = make(chan worker.WorkerExit, 32)
		runError          error
	)

	errgrp.Go(func() error {
		return app.controlWorker.Run(ctx, workerExitChannel, app.uuid, *app.coreSupervisor.GetCurrentEndpoint())
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
			switch workerExitResult.GetKind() {
			case errorhelper.ExitFatal:
				runError = workerExitResult.GetError()
				stop()
				break outer
			case errorhelper.ExitShutdown:
				stop()
				break outer
			}
		}
	}

	err := errgrp.Wait() // Wait for all go routines to finish

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

func reconnectToCore(ctx context.Context, workerExitResult worker.WorkerExit, workerExitChannel chan<- worker.WorkerExit) {
	// TODO Implement
	panic("unimplemented")
}
