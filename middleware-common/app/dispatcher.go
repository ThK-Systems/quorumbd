package app

import (
	"context"
	"log/slog"
	"net"
	"sync"

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

func (dispatcher *dispatcher) run(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	var err error
	dispatcher.conn, dispatcher.connEpoch, err = dispatcher.app.coreSupervisor.Dial(ctx)
	if err != nil {
		return err
	}
	dispatcher.logger.Info("Connected to %s with epoch %03d", dispatcher.conn.RemoteAddr().String(), dispatcher.connEpoch)

	errCh := make(chan error, 2)

	var once sync.Once
	closeConn := func() {
		once.Do(func() {
			dispatcher.conn.Close()
		})
	}

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
	closeConn()

	<-errCh
	return err
}

func sendLoop(ctx context.Context, conn net.Conn) error {
	// TODO: Implement
	panic("unimplemented")
}

func recvLoop(ctx context.Context, conn net.Conn) error {
	// TODO: Implement
	panic("unimplemented")
}
