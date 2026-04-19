// Package control provides messages and structures and handlers and common logic for the control plane
package control

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	commoncontrol "quorumbd.net/common/control"
)

type Dispatcher struct {
	logger     *slog.Logger
	toCore     chan commoncontrol.ControlMessage
	registry   map[uint32]commoncontrol.MessageHandler
	registryMu sync.RWMutex
}

var (
	dispatcherInstance *Dispatcher
	dispatcherOnce     sync.Once
)

func InitDispatcher(parentLogger *slog.Logger) *Dispatcher {
	dispatcherOnce.Do(func() {
		dispatcherInstance = &Dispatcher{
			logger:   parentLogger.With("module", "dispatcher"),
			toCore:   make(chan commoncontrol.ControlMessage, 6),
			registry: make(map[uint32]commoncontrol.MessageHandler),
		}
	})
	return dispatcherInstance
}

func GetDispatcher() (*Dispatcher, error) {
	if dispatcherInstance == nil {
		return nil, fmt.Errorf("dispatcher is not initialized")
	}
	return dispatcherInstance, nil
}

func (dispatcher *Dispatcher) SendMessageToCore(msg commoncontrol.ControlMessage) error {
	select {
	case dispatcher.toCore <- msg:
	case <-time.After(1 * time.Second): // TOCONFIG
		return fmt.Errorf("send timeout of message to core: %+v", msg)
	}
	return nil
}

func (dispatcher *Dispatcher) RegisterForCoreMessage(messageType uint32, messageHandler commoncontrol.MessageHandler) error {
	dispatcher.registryMu.Lock()
	defer dispatcher.registryMu.Unlock()
	if dispatcher.registry[messageType] != nil {
		return fmt.Errorf("message type %d is already registered to %+v", messageType, dispatcher.registry[messageType])
	}
	dispatcher.logger.Debug("Registering handler for message type", "type", messageType, "handler", fmt.Sprintf("%+v", messageHandler))
	dispatcher.registry[messageType] = messageHandler
	return nil
}

func (dispatcher *Dispatcher) UnregisterForCoreMessage(messageType uint32) error {
	dispatcher.registryMu.Lock()
	defer dispatcher.registryMu.Unlock()
	if dispatcher.registry[messageType] == nil {
		return fmt.Errorf("message type %d is not registered", messageType)
	}
	dispatcher.logger.Debug("Unregistering handler for message type", "type", messageType)
	delete(dispatcher.registry, messageType)
	return nil
}
