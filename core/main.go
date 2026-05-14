package main

import (
	"fmt"
	"os"

	"quorumbd.net/core/internal/core"
)

func main() {
	c, err := core.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if err := c.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
