// Package coreconnection manages the connection to core
package coreconnection

import (
	"thk-systems.net/quorumbd/middleware-common/config"
)

type CoreManager struct {
	config              *config.CoreConnectionConfig
	currentConnection   *CoreConnection
	primaryConnection   *CoreConnection
	fallbackConnections []*CoreConnection
}

func New(cfg *config.CoreConnectionConfig) (*CoreManager, error) {
	primary, err := FromURI(cfg.Server)
	if err != nil {
		return nil, err
	}

	fallbacks := make([]*CoreConnection, len(cfg.ServerFallback))
	for _, fallbackURI := range cfg.ServerFallback {
		fallback, err := FromURI(fallbackURI)
		if err != nil {
			return nil, err
		}
		fallbacks = append(fallbacks, fallback)
	}

	return &CoreManager{
		config:              cfg,
		currentConnection:   nil,
		primaryConnection:   primary,
		fallbackConnections: fallbacks,
	}, nil
}

func (cm *CoreManager) IsConnected() bool {
	return cm.currentConnection != nil
}
