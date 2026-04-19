// Package coreconnection manages the connection to core
package coreconnection

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"quorumbd.net/middleware-common/config"
)

type CoreSupervisor struct {
	config            *config.CoreConnectionConfig
	logger            *slog.Logger
	connectionEpoch   atomic.Uint32
	currentEndpoint   atomic.Pointer[CoreEndpoint]
	primaryEndpoint   *CoreEndpoint
	fbeMutex          sync.RWMutex
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

func (cs *CoreSupervisor) GetCurrentEndpoint() *CoreEndpoint {
	return cs.currentEndpoint.Load()
}

func (cs *CoreSupervisor) GetConnectionEpoch() uint32 {
	return cs.connectionEpoch.Load()
}

func (cs *CoreSupervisor) IsConnected() bool {
	return cs.currentEndpoint.Load() != nil // TODO - Check for connection not endpoint
}

func (cs *CoreSupervisor) IsPrimary() bool {
	return cs.IsConnected() && cs.currentEndpoint.Load() == cs.primaryEndpoint
}

func (cs *CoreSupervisor) Try(ctx context.Context, initialBackoff time.Duration, maxBackoff time.Duration, probeInfinitely bool) error {
	return cs.Retry(ctx, initialBackoff, maxBackoff, probeInfinitely, false, nil)
}

func (cs *CoreSupervisor) RetryPrimary(ctx context.Context, initialBackoff time.Duration, maxBackoff time.Duration) error {
	return cs.Retry(ctx, initialBackoff, maxBackoff, true, true, nil)
}

func (cs *CoreSupervisor) Retry(ctx context.Context, initialBackoff time.Duration, maxBackoff time.Duration, probeInfinitely bool, primaryOnly bool, endpointToExclude *CoreEndpoint) error {

	exclude := "none"
	if endpointToExclude != nil {
		exclude = endpointToExclude.toURI()
	}
	cs.logger.Info("Starting core probe", "excluding", exclude, "primary_only", primaryOnly, "infinitely", probeInfinitely)

	var err error
	backoff := min(initialBackoff, maxBackoff)

	for {
		// wait or shutdown
		tryCount := 0
		cs.logger.Info("Probing core endpoint(s)", "retry_in", backoff.String())
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}

		// try primary
		if endpointToExclude != cs.primaryEndpoint {
			tryCount++
			cs.logger.Debug("Probing primary core endpoint", "address", cs.primaryEndpoint.toURI())
			err = cs.primaryEndpoint.tryDial(ctx)
			if err == nil {
				cs.logger.Info("Primary core endpoint is reachable", "address", cs.primaryEndpoint.toURI())
				cs.setNewCurrentEndpoint(cs.primaryEndpoint)
				return nil
			}
		}

		// try fallbacks
		if !primaryOnly {
			cs.fbeMutex.RLock()
			fallbacks := cs.fallbackEndpoints
			cs.fbeMutex.RUnlock()
			for _, fb := range fallbacks {
				if fb == endpointToExclude {
					break
				}
				tryCount++
				cs.logger.Debug("Probing fallback core connection", "address", fb.toURI())
				err = fb.tryDial(ctx)
				if err == nil {
					cs.logger.Info("Fallback core endpoint is reachable", "address", fb.toURI())
					cs.setNewCurrentEndpoint(fb)
					return nil
				}
			}
		}

		// check for no available endpoints
		if tryCount == 0 {
			return fmt.Errorf("no core endpoints available")
		}

		// increment backoff
		if backoff < maxBackoff {
			backoff = min(max(backoff, 1*time.Second)*2, maxBackoff)
		} else if !probeInfinitely {
			return fmt.Errorf("core endpoints not reachable")
		}
	}
}

func (cs *CoreSupervisor) setNewCurrentEndpoint(endpoint *CoreEndpoint) {
	cs.currentEndpoint.Store(endpoint)
	cs.connectionEpoch.Add(1)
}
