package app

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"sync/atomic"
	"time"

	"quorumbd.net/common/control"
	commonio "quorumbd.net/common/io"
)

type dispatcher struct {
	app       *App
	logger    *slog.Logger
	conn      net.Conn
	connEpoch atomic.Uint32
	toCore    <-chan control.ControlMessage
	fromCore  chan<- control.ControlMessage
}

func newDispatcher(app *App) *dispatcher {
	return &dispatcher{
		app:      app,
		logger:   app.logger.With("module", "dispatcher"),
		toCore:   make(chan control.ControlMessage, 1),
		fromCore: make(chan control.ControlMessage, 1),
	}
}

func (dispatcher *dispatcher) restartOnCoreReconnect() bool {
	return true
}

func (dispatcher *dispatcher) getCoreConnectionEpoch() uint32 {
	return uint32(dispatcher.connEpoch.Load())
}

func (dispatcher *dispatcher) run(parentCtx context.Context, workerExitCh chan<- WorkerExit) error {
	childContext, cancel := context.WithCancel(parentCtx)
	defer cancel()

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

	if err := commonio.WriteFull(dispatcher.conn, append([]byte("CTRL"), dispatcher.app.uuid[:]...)); err != nil {
		return err
	}

	errCh := make(chan error, 2)

	go func() {
		dispatcher.logger.Info("Starting receive loop")
		errCh <- recvLoop(childContext, dispatcher.conn, dispatcher.fromCore)
	}()

	go func() {
		dispatcher.logger.Info("Starting send loop")
		errCh <- sendLoop(childContext, dispatcher.conn, dispatcher.toCore)
	}()

	err = <-errCh // Get error from first loop
	cancel()
	dispatcher.conn.Close()
	<-errCh // Waiting for second loop (ignoring error) in cause of `cancel()`

	workerExit := newWorkerExit(dispatcher, err)
	workerExitCh <- workerExit

	dispatcher.logger.Error("dispatcher exit", "kind", workerExit.kind.String(), "err", workerExit.err)

	return err
}

func sendLoop(ctx context.Context, conn net.Conn, toCore <-chan control.ControlMessage) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-toCore:
			if !ok {
				return nil // channel closed
			}
			if err := send(conn, msg, 3*time.Second); err != nil {
				return err
			}
		}
	}
}

func send(conn net.Conn, msg control.ControlMessage, timeout time.Duration) error {
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

func recvLoop(ctx context.Context, conn net.Conn, fromCore chan<- control.ControlMessage) error {
	for {
		msg, err := receive(conn)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case fromCore <- msg:
		}
	}
}

func receive(conn net.Conn) (control.ControlMessage, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(conn, lenBuf[:]); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lenBuf[:])

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

	msg, err := control.NewMessage(head.Type)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}

	return msg, nil
}
