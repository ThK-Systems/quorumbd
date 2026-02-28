// Main package of middleware-qemu-nbd
package main

import (
	"fmt"
	"os"

	logging "thk-systems.net/quorumbd/common/logging"
	app "thk-systems.net/quorumbd/middleware-common/app"
	config "thk-systems.net/quorumbd/middleware-qemu-nbd/internal/config"
	implementation "thk-systems.net/quorumbd/middleware-qemu-nbd/internal/implementation"
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

	return app.New(implementation.New(), config.Get().ToMiddlewareConfig()).Run()
}
