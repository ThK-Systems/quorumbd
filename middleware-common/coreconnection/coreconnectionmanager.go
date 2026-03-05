// Package coreconnection manages the connection to core
package coreconnection

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"thk-systems.net/quorumbd/middleware-common/config"
)

type CoreConnectionManager struct {
	config              *config.CoreConnectionConfig
	logger              *slog.Logger
	mutex               sync.RWMutex
	currentConnection   atomic.Pointer[CoreConnection]
	primaryConnection   *CoreConnection
	fallbackConnections []*CoreConnection
}

func New(cfg *config.CoreConnectionConfig, logger *slog.Logger) (*CoreConnectionManager, error) {
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

	return &CoreConnectionManager{
		config:              cfg,
		logger:              logger.With("module", "coremanager"),
		primaryConnection:   primary,
		fallbackConnections: fallbacks,
	}, nil
}

func (cm *CoreConnectionManager) Probe(ctx context.Context, initialBackoff time.Duration, maxBackoff time.Duration, probeInfinitely bool, primaryOnly bool) error {

	cm.logger.Info("Starting core probe")

	var (
		err  error
		conn *CoreConnection
	)
	backoff := min(initialBackoff, maxBackoff)

	for {
		// 1. wait or shutdown
		cm.logger.Info("Probing core connection(s)", "retry_in", backoff.String())
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		// 2. try primary
		cm.logger.Debug("Probing primary core connection", "address", cm.primaryConnection.toURI())
		conn, err = cm.primaryConnection.tryDial(ctx)
		if err == nil {
			cm.logger.Info("Primary core connection is reachable", "address", cm.primaryConnection.toURI())
			cm.currentConnection.Store(conn)
			return nil
		}

		// 3. try fallbacks
		if !primaryOnly {
			for _, fb := range cm.fallbackConnections {
				cm.logger.Debug("Probing fallback core connection", "address", fb.toURI())
				conn, err = fb.tryDial(ctx)
				if err == nil {
					cm.logger.Info("Fallback core connection is reachable", "address", fb.toURI())
					cm.currentConnection.Store(conn)
					return nil
				}
			}
		}

		// 4. increment backoff
		if backoff < maxBackoff {
			backoff = min(max(backoff, 1*time.Second)*2, maxBackoff)
		} else if !probeInfinitely {
			return fmt.Errorf("core connections not reachable")
		}
	}
}

func (cm *CoreConnectionManager) IsConnected() bool {
	return cm.currentConnection.Load() != nil
}

func (cm *CoreConnectionManager) IsPrimary() bool {
	return cm.IsConnected() && cm.currentConnection.Load() == cm.primaryConnection
}
