// Main package of middleware-qemu-nbd
package main

import (
	"fmt"
	"os"

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

	// read state

	// register at core

	// create nbd server socket and listen
	// TODO: NEXT

	// open data sockets and start go routines (tcp/udp)

	// open control socket and start go routine (tcp/udp)

	logger.Info("quorumbd qemu-nbd-server is completely started")

	return nil
}
