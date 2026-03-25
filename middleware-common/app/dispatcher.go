package app

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"quorumbd.net/common/control"
	commonio "quorumbd.net/common/io"
)

const (
	maxFrameSize = 1 << 20 // 1MB
)

var dispatcherSingleton struct {
	mu          sync.Mutex
	initialized bool
}

type dispatcher struct {
	app        *App
	logger     *slog.Logger
	conn       net.Conn
	connEpoch  atomic.Uint32
	toCore     chan control.ControlMessage
	registry   map[uint32]control.MessageHandler
	registryMu sync.RWMutex
}

func newDispatcher(app *App) *dispatcher {
	dispatcherSingleton.mu.Lock()
	defer dispatcherSingleton.mu.Unlock()

	if dispatcherSingleton.initialized {
		panic("dispatcher singleton already instantiated")
	}

	dispatcherSingleton.initialized = true

	return &dispatcher{
		app:      app,
		logger:   app.logger.With("module", "dispatcher"),
		toCore:   make(chan control.ControlMessage, 6),
		registry: make(map[uint32]control.MessageHandler),
	}
}

func (dispatcher *dispatcher) restartOnCoreReconnect() bool {
	return true
}

func (dispatcher *dispatcher) getCoreConnectionEpoch() uint32 {
	return uint32(dispatcher.connEpoch.Load())
}

func (dispatcher *dispatcher) SendMessageToCore(msg control.ControlMessage) error {
	select {
	case dispatcher.toCore <- msg:
	case <-time.After(1 * time.Second): // TOCONFIG
		return fmt.Errorf("send timeout of message to core: %+v", msg)
	}
	return nil
}

func (dispatcher *dispatcher) RegisterForCoreMessage(messageType uint32, messageHandler control.MessageHandler) error {
	dispatcher.registryMu.Lock()
	defer dispatcher.registryMu.Unlock()
	if dispatcher.registry[messageType] != nil {
		return fmt.Errorf("message type %d is already registered to %+v", messageType, dispatcher.registry[messageType])
	}
	dispatcher.logger.Debug("Registering handler for message type", "type", messageType, "handler", fmt.Sprintf("%+v", messageHandler))
	dispatcher.registry[messageType] = messageHandler
	return nil
}

func (dispatcher *dispatcher) run(parentCtx context.Context, workerExitCh chan<- WorkerExit) error {
	childContext, cancel := context.WithCancel(parentCtx)
	defer cancel()

	dispatcher.logger.Info("Starting dispatcher")

	var (
		err       error
		connEpoch uint32
	)
	dispatcher.conn, connEpoch, err = dispatcher.app.coreSupervisor.Dial(childContext)
	dispatcher.connEpoch.Store(connEpoch)
	if err != nil {
		return err
	}
	defer dispatcher.conn.Close()
	dispatcher.logger.Info("Connected", "address", dispatcher.conn.RemoteAddr().String(), "epoch", dispatcher.connEpoch.Load())

	go func() {
		<-childContext.Done()
		dispatcher.conn.Close()
	}()

	if err := commonio.WriteFull(dispatcher.conn, append([]byte("CTRL"), dispatcher.app.uuid[:]...)); err != nil {
		return err
	}

	errCh := make(chan error, 2)

	go func() {
		dispatcher.logger.Info("Starting receive loop")
		errCh <- dispatcher.recvLoop(childContext)
	}()

	go func() {
		dispatcher.logger.Info("Starting send loop")
		errCh <- dispatcher.sendLoop(childContext)
	}()

	err = <-errCh // Get error from first loop
	cancel()
	if err != nil {
		dispatcher.logger.Error("Closing connection", "error", err)
	} else {
		dispatcher.logger.Info("Closing connection because context done")
	}
	dispatcher.conn.Close()
	<-errCh // Waiting for second loop (ignoring error) in cause of `cancel()`

	workerExit := newWorkerExit(dispatcher, err)
	workerExitCh <- workerExit

	dispatcher.logger.Info("Dispatcher exit", "kind", workerExit.kind.String(), "err", workerExit.err)

	return err
}

func (dispatcher *dispatcher) sendLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			dispatcher.logger.Info("Stopping send loop because context done")
			return nil
		case msg, ok := <-dispatcher.toCore:
			if !ok {
				dispatcher.logger.Warn("Stopping send loop because of closed channel")
				return nil // channel closed
			}
			dispatcher.logger.Debug("Send control message to core", "message", fmt.Sprintf("%+v", msg))
			if err := dispatcher.send(msg, 3*time.Second); err != nil { // TOCONFIG
				if ctx.Err() != nil {
					dispatcher.logger.Info("Stopping send loop because context done")
					return nil
				}
				dispatcher.logger.Warn("Stopping send loop", "error", err)
				return err
			}
		}
	}
}

func (dispatcher *dispatcher) send(msg control.ControlMessage, timeout time.Duration) error {
	if err := dispatcher.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	defer dispatcher.conn.SetWriteDeadline(time.Time{})

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(data)))

	if err := commonio.WriteFull(dispatcher.conn, header[:]); err != nil {
		return err
	}
	if err := commonio.WriteFull(dispatcher.conn, data); err != nil {
		return err
	}

	return nil
}

func (dispatcher *dispatcher) recvLoop(ctx context.Context) error {
	for {
		msg, err := dispatcher.receive()
		if err != nil {
			if ctx.Err() != nil {
				dispatcher.logger.Info("Stopping receive loop because context done")
				return nil
			}
			dispatcher.logger.Warn("Stopping receive loop", "error", err)
			return err
		}

		dispatcher.logger.Debug("Received control message from core", "message", fmt.Sprintf("%+v", msg))

		dispatcher.registryMu.RLock()
		handler := dispatcher.registry[msg.Type()]
		dispatcher.registryMu.RUnlock()

		if handler == nil {
			dispatcher.logger.Warn("No handler for message type", "type", msg.Type())
			continue
		}

		handler.HandleMessage(ctx, msg)
	}
}

func (dispatcher *dispatcher) receive() (control.ControlMessage, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(dispatcher.conn, lenBuf[:]); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lenBuf[:])
	if length > maxFrameSize {
		return nil, fmt.Errorf("frame too large: %d > %d", length, maxFrameSize)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(dispatcher.conn, data); err != nil {
		return nil, err
	}

	var head struct {
		Type uint32 `json:"type"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return nil, err
	}

	msg, err := control.NewMessage(head.Type)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}

	return msg, nil
}
