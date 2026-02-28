// Package implementation implements the adaptor interface of middleware-common
package implementation

import (
	"errors"
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

// GetImplementationName is an interface method of common-middleware.Adapter
func (impl *Implementation) GetImplementationName() string {
	return "qemu-nbd"
}

// Listen is an interface method of common-middleware.Adapter
func (impl *Implementation) Listen() error {
	return errors.New("not implemented")
}

// Disconnect is an interface method of common-middleware.Adapter
func (impl *Implementation) Disconnect() error {
	return errors.New("not implemented")
}
