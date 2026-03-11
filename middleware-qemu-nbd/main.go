// Main package of middleware-qemu-nbd
package main

import (
	"fmt"
	"os"

	logging "quorumbd.net/common/logging"
	app "quorumbd.net/middleware-common/app"
	config "quorumbd.net/middleware-qemu-nbd/internal/config"
	implementation "quorumbd.net/middleware-qemu-nbd/internal/implementation"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
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
	err = logging.Initialize(config.Get().LoggingConfig)
	if err != nil {
		return err
	}

	app, err := app.New(
		implementation.New(config.Get(), logging.GetDefaultLogger()),
		config.Get().ToMiddlewareConfig(),
		logging.GetDefaultLogger(),
	)
	if err != nil {
		return err
	}

	if err := app.Run(); err != nil {
		return err
	}

	// terminate logging
	return logging.CloseLogging()
}
