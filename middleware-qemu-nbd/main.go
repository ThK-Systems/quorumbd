// Main package of middleware-qemu-nbd
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"thk-systems.net/quorumbd/common/logging"
	"thk-systems.net/quorumbd/middleware-qemu-nbd/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// load and init config
	err := config.Load()
	if err != nil {
		return err
	}

	// init logging
	err = logging.Initialize(config.Get().Logging)
	if err != nil {
		return err
	}

	logger := logging.For("main")
	logger.Info("quorumbd qemu-nbd-server is about to start ...")

	// prepare shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// channel to receive fatal errors from components
	errCh := make(chan error, 1)

	// read state

	// register at core

	// create nbd server socket and listen
	// TODO: NEXT

	// open data sockets and start go routines

	// open control socket and start go routine

	logger.Info("quorumbd qemu-nbd-server is completely started")

	// orchestrator
	select {
	case <-ctx.Done():
		logger.Info("quorumbd qemu-nbd-server stopped because of shutdown signal")
		return nil

	case err := <-errCh:
		logger.Error("quorumbd qemu-nbd-server stopped because component failed", "err", err)
		return err
	}
}
