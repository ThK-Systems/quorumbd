package app

import (
	"context"
	"log/slog"
	"net"

	"quorumbd.net/common/control"
)

type dispatcher struct {
	app       *App
	logger    *slog.Logger
	conn      net.Conn
	connEpoch uint32
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

func (dispatcher *dispatcher) run(parentCtx context.Context, workerExitCh chan<- WorkerExit) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	var err error
	dispatcher.conn, dispatcher.connEpoch, err = dispatcher.app.coreSupervisor.Dial(ctx)
	if err != nil {
		return err
	}
	defer dispatcher.conn.Close()
	dispatcher.logger.Info("Connected", "address", dispatcher.conn.RemoteAddr().String(), "epoch", dispatcher.connEpoch)

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
