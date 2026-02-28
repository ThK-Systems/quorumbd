// Package implementation implements the adaptor interface of middleware-common
package implementation

import (
	"log/slog"

	"thk-systems.net/quorumbd/middleware-qemu-nbd/internal/config"
)

type Implementation struct {
	Config *config.Config
	Logger *slog.Logger
}

func New(config *config.Config, logger *slog.Logger) *Implementation {
	return &Implementation{
		Config: config,
		Logger: logger,
	}
}

// GetImplementationName is an interface method
func (impl *Implementation) GetImplementationName() string {
	return "qemu-nbd"
}

// Listen is an interface method
func (impl *Implementation) Listen() error {
	return nil
}

// Connect is an interface method
func (impl *Implementation) Connect() error {
	return nil
}

// Disconnect is an interface method
func (impl *Implementation) Disconnect() error {
	return nil
}
