package main

import (
	"fmt"
	"log/slog"
	"os"

	"thk-systems.net/quorumbd/middleware-qemu-nbd/internal/config"
	"thk-systems.net/quorumbd/middleware-qemu-nbd/internal/logging"
)

func main() {
	// Load and init config
	err := config.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	// Init logging
	err = logging.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	slog.Info("QuorumBD qemu-nbd middleware is about to start ...")

	// Read state

	// register at core

	// open data sockets and start go routines

	// open control socket and start go routine

	// create nbd server socket and listen

	slog.Info("QuorumBD qemu-nbd is completely started")

}
