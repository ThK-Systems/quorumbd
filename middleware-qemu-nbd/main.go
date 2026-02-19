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

	slog.Info("qemu-nbd middleware starting ...")

	// Read state
	// ...
}
