// Package coreconnection manages the connection to core
package coreconnection

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"thk-systems.net/quorumbd/middleware-common/config"
)

type CoreManager struct {
	config              *config.CoreConnectionConfig
	logger              *slog.Logger
	currentConnection   *CoreConnection
	primaryConnection   *CoreConnection
	fallbackConnections []*CoreConnection
}

func New(cfg *config.CoreConnectionConfig, logger *slog.Logger) (*CoreManager, error) {
	primary, err := fromURI(cfg.Server)
	if err != nil {
		return nil, err
	}

	fallbacks := make([]*CoreConnection, 0, len(cfg.ServerFallback))
	for _, fallbackURI := range cfg.ServerFallback {
		fallback, err := fromURI(fallbackURI)
		if err != nil {
			return nil, err
		}
		fallbacks = append(fallbacks, fallback)
	}

	return &CoreManager{
		config:              cfg,
		logger:              logger.With("module", "coremanager"),
		currentConnection:   nil,
		primaryConnection:   primary,
		fallbackConnections: fallbacks,
	}, nil
}

func (cm *CoreManager) Probe(ctx context.Context, initialBackoff time.Duration, maxBackoff time.Duration, probeInfinitely bool) error {

	cm.logger.Info("Starting core probe")

	var err error
	backoff := min(initialBackoff, maxBackoff)

	for {
		// 1. wait or shutdown
		cm.logger.Info("Probing core connections", "retry_in", backoff.String())
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		// 2. try primary
		cm.logger.Debug("Probing primary core connection", "address", cm.primaryConnection.toURI())
		cm.currentConnection, err = cm.primaryConnection.tryDial(ctx)
		if err == nil {
			cm.logger.Info("Primary core connection is reachable", "address", cm.primaryConnection.toURI())
			return nil
		}

		// 3. try fallbacks
		for _, fb := range cm.fallbackConnections {
			cm.logger.Debug("Probing fallback core connection", "address", fb.toURI())
			cm.currentConnection, err = fb.tryDial(ctx)
			if err == nil {
				cm.logger.Info("Fallback core connection is reachable", "address", fb.toURI())
				return nil
			}
		}

		// 4. increment backoff
		if backoff < maxBackoff {
			backoff = min(max(backoff, 1*time.Second)*2, maxBackoff)
		} else if !probeInfinitely {
			return fmt.Errorf("no reachable core connection (primary + %d fallbacks)", len(cm.fallbackConnections))
		}
	}
}

func (cm *CoreManager) IsConnected() bool {
	return cm.currentConnection != nil
}

func (cm *CoreManager) IsPrimary() bool {
	return cm.IsConnected() && cm.currentConnection == cm.primaryConnection
}
