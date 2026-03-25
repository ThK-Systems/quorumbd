package main

import (
	"fmt"
	"os"

	"quorumbd.net/core/internal/config"
)

func main() {
	core, err := newCore()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	core.run()
}

type core struct {
	config *config.Config
}

func newCore() (*core, error) {
	config, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &core{
		config: config,
	}, nil
}

func (c *core) run() {
	panic("unimplemented")
}
