// Package coreconnection manages the connection to core
package coreconnection

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"thk-systems.net/quorumbd/middleware-common/config"
)

type CoreSupervisor struct {
	config            *config.CoreConnectionConfig
	logger            *slog.Logger
	currentEndpoint   atomic.Pointer[CoreEndpoint]
	primaryEndpoint   *CoreEndpoint
	fallbackEndpoints []*CoreEndpoint
}

func New(cfg *config.CoreConnectionConfig, logger *slog.Logger) (*CoreSupervisor, error) {
	primary, err := fromURI(cfg.Server)
	if err != nil {
		return nil, err
	}

	fallbacks := make([]*CoreEndpoint, 0, len(cfg.ServerFallback))
	for _, fallbackURI := range cfg.ServerFallback {
		fallback, err := fromURI(fallbackURI)
		if err != nil {
			return nil, err
		}
		fallbacks = append(fallbacks, fallback)
	}

	return &CoreSupervisor{
		config:            cfg,
		logger:            logger.With("module", "coresupervisor"),
		primaryEndpoint:   primary,
		fallbackEndpoints: fallbacks,
	}, nil
}

func (cs *CoreSupervisor) Probe(ctx context.Context, initialBackoff time.Duration, maxBackoff time.Duration, probeInfinitely bool, primaryOnly bool) (*CoreEndpoint, error) {
	cs.logger.Info("Starting core probe")

	var (
		err  error
		conn *CoreEndpoint
	)
	backoff := min(initialBackoff, maxBackoff)

	for {
		// 1. wait or shutdown
		cs.logger.Info("Probing core endpoint(s)", "retry_in", backoff.String())
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}

		// 2. try primary
		cs.logger.Debug("Probing primary core endpoint", "address", cs.primaryEndpoint.toURI())
		conn, err = cs.primaryEndpoint.tryDial(ctx)
		if err == nil {
			cs.logger.Info("Primary core endpoint is reachable", "address", cs.primaryEndpoint.toURI())
			return conn, nil
		}

		// 3. try fallbacks
		if !primaryOnly {
			for _, fb := range cs.fallbackEndpoints {
				cs.logger.Debug("Probing fallback core connection", "address", fb.toURI())
				conn, err = fb.tryDial(ctx)
				if err == nil {
					cs.logger.Info("Fallback core endpoint is reachable", "address", fb.toURI())
					return conn, nil
				}
			}
		}

		// 4. increment backoff
		if backoff < maxBackoff {
			backoff = min(max(backoff, 1*time.Second)*2, maxBackoff)
		} else if !probeInfinitely {
			return nil, fmt.Errorf("core endpoints not reachable")
		}
	}
}

func (cs *CoreSupervisor) IsConnected() bool {
	return cs.currentEndpoint.Load() != nil // TODO - Check for connection not endpoint
}

func (cs *CoreSupervisor) IsPrimary() bool {
	return cs.IsConnected() && cs.currentEndpoint.Load() == cs.primaryEndpoint
}
