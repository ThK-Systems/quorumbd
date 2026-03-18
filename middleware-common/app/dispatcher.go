package app

import (
	"context"
	"log/slog"
	"net"
	"sync/atomic"

	"quorumbd.net/common/control"
)

type dispatcher struct {
	app       *App
	logger    *slog.Logger
	conn      net.Conn
	connEpoch atomic.Uint32
	toCore    chan<- control.ControlMessage
	fromCore  <-chan control.ControlMessage
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
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	var (
		err error
		connEpoch uint32
	)
	dispatcher.conn, connEpoch, err = dispatcher.app.coreSupervisor.Dial(ctx)
	dispatcher.connEpoch.Store(connEpoch)
	if err != nil {
		return err
	}
	defer dispatcher.conn.Close()
	dispatcher.logger.Info("Connected", "address", dispatcher.conn.RemoteAddr().String(), "epoch", dispatcher.connEpoch.Load())

	errCh := make(chan error, 2)

	go func() {
		dispatcher.logger.Info("Starting receive loop")
		errCh <- recvLoop(ctx, dispatcher.conn)
	}()

	go func() {
		dispatcher.logger.Info("Starting send loop")
		errCh <- sendLoop(ctx, dispatcher.conn)
	}()

	err = <-errCh
	cancel()
	<-errCh

	workerExit := newWorkerExit(dispatcher, err)
	workerExitCh <- workerExit

	dispatcher.logger.Error("dispatcher exit", "kind", workerExit.kind.String(), "err", workerExit.err)

	return err
}

func sendLoop(ctx context.Context, conn net.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		}
	}
}

func recvLoop(ctx context.Context, conn net.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		}
	}
}
