// Main package of middleware-qemu-nbd
package main

import (
	"fmt"
	"os"

	"thk-systems.net/quorumbd/common/logging"
	"thk-systems.net/quorumbd/middleware-qemu-nbd/internal/config"
)

func main() {
	// Load and init config
	err := config.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	// Init logging
	err = logging.Init(config.Get().Logging)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger := logging.For("main")

	logger.Info("QuorumBD qemu-nbd middleware is about to start ...")

	// Read state

	// register at core

	// open data sockets and start go routines

	// open control socket and start go routine

	// create nbd server socket and listen

	logger.Info("QuorumBD qemu-nbd is completely started")

}
