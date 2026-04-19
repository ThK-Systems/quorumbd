package control

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"

	commoncontrol "quorumbd.net/common/control"
	commonio "quorumbd.net/common/io"

	"quorumbd.net/middleware-common/coreconnection"
	"quorumbd.net/middleware-common/worker"
)

const (
	maxFrameSize = 1 << 20 // 1MB
)

var controlWorkerSingleton struct {
	mu          sync.Mutex
	initialized bool
}

type ControlWorker struct {
	logger     *slog.Logger
	dispatcher *Dispatcher
}

func NewControlWorker(parentLogger *slog.Logger, dispatcher *Dispatcher) *ControlWorker {
	controlWorkerSingleton.mu.Lock()
	defer controlWorkerSingleton.mu.Unlock()

	if controlWorkerSingleton.initialized {
		panic("controlworker singleton already instantiated")
	}

	controlWorkerSingleton.initialized = true

	return &ControlWorker{
		logger:     parentLogger.With("module", "controlworker"),
		dispatcher: dispatcher,
	}
}

func (cw *ControlWorker) String() string {
	return "controlworker"
}

func (cw *ControlWorker) RestartOnCoreReconnect() bool {
	return true
}

func (cw *ControlWorker) Run(parentCtx context.Context, workerExitCh chan<- worker.WorkerExit, middlewareUUID uuid.UUID, coreEndpoint coreconnection.CoreEndpoint) {
	childContext, cancel := context.WithCancel(parentCtx)
	defer cancel()

	cw.logger.Info("Starting control worker")

	conn, err := coreEndpoint.Dial(childContext)
	if err != nil {
		cw.exit(err, workerExitCh)
		return
	}
	defer conn.Close()
	cw.logger.Info("Connected", "address", conn.RemoteAddr().String())

	go func() {
		<-childContext.Done()
		conn.Close()
	}()

	if err := commonio.WriteFull(conn, append([]byte("CTRL"), middlewareUUID[:]...)); err != nil {
		cw.exit(err, workerExitCh)
		return
	}

	errCh := make(chan error, 2)

	go func() {
		cw.logger.Info("Starting receive loop")
		errCh <- cw.recvLoop(childContext, conn)
	}()

	go func() {
		cw.logger.Info("Starting send loop")
		errCh <- cw.sendLoop(childContext, conn)
	}()

	err = <-errCh // Get error from first loop
	shutdownRequested := parentCtx.Err() != nil
	cancel()
	if shutdownRequested && err != nil {
		cw.logger.Info("Closing connection because context done")
		err = nil
	} else if err != nil {
		cw.logger.Error("Closing connection", "error", err)
	} else {
		cw.logger.Info("Closing connection because context done")
	}
	conn.Close()
	<-errCh // Waiting for second loop (ignoring error) in cause of `cancel()`

	cw.exit(err, workerExitCh)
}

func (cw *ControlWorker) exit(err error, workerExitCh chan<- worker.WorkerExit) {
	workerExit := worker.NewWorkerExit(cw, err)
	workerExitCh <- workerExit
	cw.logger.Info("Dispatcher exit: " + workerExit.String())
}

func (cw *ControlWorker) sendLoop(ctx context.Context, conn net.Conn) error {
	for {
		select {
		case <-ctx.Done():
			cw.logger.Info("Stopping send loop because context done")
			return nil
		case msg, ok := <-cw.dispatcher.toCore:
			if !ok {
				cw.logger.Warn("Stopping send loop because of closed channel")
				return nil // channel closed
			}
			cw.logger.Debug("Send control message to core", "message", fmt.Sprintf("%+v", msg))
			if err := cw.send(msg, 3*time.Second, conn); err != nil { // TOCONFIG
				if ctx.Err() != nil {
					cw.logger.Info("Stopping send loop because context done")
					return nil
				}
				cw.logger.Warn("Stopping send loop", "error", err)
				return err
			}
		}
	}
}

func (cw *ControlWorker) send(msg commoncontrol.ControlMessage, timeout time.Duration, conn net.Conn) error {
	if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	defer conn.SetWriteDeadline(time.Time{})

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(data)))

	if err := commonio.WriteFull(conn, header[:]); err != nil {
		return err
	}
	if err := commonio.WriteFull(conn, data); err != nil {
		return err
	}

	return nil
}

func (cw *ControlWorker) recvLoop(ctx context.Context, conn net.Conn) error {
	for {
		msg, err := cw.receive(conn)
		if err != nil {
			if ctx.Err() != nil {
				cw.logger.Info("Stopping receive loop because context done")
				return nil
			}
			cw.logger.Warn("Stopping receive loop", "error", err)
			return err
		}

		cw.logger.Debug("Received control message from core", "message", fmt.Sprintf("%+v", msg))

		handler := cw.dispatcher.getHandlerForMessageType(msg.Type())
		if handler == nil {
			cw.logger.Warn("No handler for message type", "type", msg.Type())
			continue
		}

		handler.HandleMessageBlocking(ctx, msg)
	}
}

func (cw *ControlWorker) receive(conn net.Conn) (commoncontrol.ControlMessage, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(conn, lenBuf[:]); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lenBuf[:])
	if length > maxFrameSize {
		return nil, fmt.Errorf("frame too large: %d > %d", length, maxFrameSize)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	var head struct {
		Type uint32 `json:"type"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return nil, err
	}

	msg, err := commoncontrol.NewMessage(head.Type)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}

	return msg, nil
}
