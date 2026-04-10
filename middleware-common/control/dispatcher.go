// Package control
package control

import (
	"log/slog"
	"sync"

	commoncontrol "quorumbd.net/common/control"
)

type Dispatcher struct {
	logger     *slog.Logger
	registry   map[uint32]commoncontrol.MessageHandler
	registryMu sync.RWMutex
}

var (
	dispatcherInstance *Dispatcher
	dispatcherOnce     sync.Once
)

func InitDispatcher(logger *slog.Logger) *Dispatcher {
	dispatcherOnce.Do(func() {
		dispatcherInstance = &Dispatcher{
			logger:   logger.With("module", "dispatcher"),
			registry: make(map[uint32]commoncontrol.MessageHandler),
		}
	})
	return dispatcherInstance
}

func GetDispatcher() *Dispatcher {
	if dispatcherInstance != nil {
		return dispatcherInstance
	}
	panic("dispatcher is not initialized")
}
