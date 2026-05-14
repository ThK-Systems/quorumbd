// Package core provides the core runtime.
package core

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/google/uuid"

	"quorumbd.net/common/logging"
	"quorumbd.net/common/state"
	"quorumbd.net/core/internal/config"
)

type Core struct {
	config *config.Config
	logger *slog.Logger
	state  *state.State
	uuid   uuid.UUID
}

func New() (*Core, error) {
	config, err := config.Load()
	if err != nil {
		return nil, err
	}

	if err := logging.Initialize(config.LoggingConfig); err != nil {
		return nil, err
	}

	logger := logging.GetDefaultLogger().With("module", "core")
	if err := state.Initialize(config.CommonConfig.StateDir, "core", logger); err != nil {
		return nil, err
	}

	coreState := state.Get()
	coreUUID, err := state.GetOrCreateUUID(coreState)
	if err != nil {
		return nil, err
	}

	logger.Info("Core initialized with uuid", "uuid", coreUUID.String())

	return &Core{
		config: config,
		logger: logger,
		state:  coreState,
		uuid:   coreUUID,
	}, nil
}

func (c *Core) Run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	listeners := make([]net.Listener, 0, len(c.config.CoreConfig.Listen))
	var listenerWG sync.WaitGroup
	var connectionWG sync.WaitGroup

	for _, uri := range c.config.CoreConfig.Listen {
		network, address, err := parseListenURI(uri)
		if err != nil {
			c.logger.Error("Core listener config invalid", "uri", uri, "error", err)
			continue
		}

		if network == "unix" {
			if err := removeUnixSocket(address); err != nil {
				c.logger.Error("Core listener cleanup failed", "network", network, "address", address, "error", err)
				continue
			}
		}

		listener, err := net.Listen(network, address) // tcp
		if err != nil {
			c.logger.Error("Core listener start failed", "network", network, "address", address, "error", err)
			continue
		}

		c.logger.Info("Core listener started", "network", network, "address", address)
		listeners = append(listeners, listener)

		listenerWG.Add(1)
		go func(listener net.Listener) {
			defer listenerWG.Done()
			for {
				conn, err := listener.Accept()
				if err != nil {
					if ctx.Err() == nil {
						c.logger.Error(
							"Core listener stopped with error",
							"network", listener.Addr().Network(),
							"address", listener.Addr().String(),
							"error", err,
						)
					}
					return
				}

				connectionWG.Add(1)
				go func(conn net.Conn) {
					defer connectionWG.Done()
					c.handleConnection(conn)
				}(conn)
			}
		}(listener)
	}

	c.logger.Info("Core listeners initialized", "started", len(listeners), "configured", len(c.config.CoreConfig.Listen))

	if len(listeners) == 0 {
		return fmt.Errorf("no core listeners started")
	}

	<-ctx.Done()
	closeListeners(listeners)
	listenerWG.Wait()
	connectionWG.Wait()

	c.logger.Info("Core is exiting ...")
	return nil
}

func (c *Core) handleConnection(conn net.Conn) {
	remote := conn.RemoteAddr().String()

	defer func() {
		_ = conn.Close()
		c.logger.Info("Core connection closed", "remote", remote)
	}()

	c.logger.Info("Core accepted connection", "remote", remote)
}

func parseListenURI(uri string) (string, string, error) {
	switch {
	case strings.HasPrefix(uri, "unix://"):
		return "unix", strings.TrimPrefix(uri, "unix://"), nil
	case strings.HasPrefix(uri, "tcp://"):
		return "tcp", strings.TrimPrefix(uri, "tcp://"), nil
	default:
		return "", "", fmt.Errorf("unsupported listen URI %q", uri)
	}
}

func closeListeners(listeners []net.Listener) {
	for _, listener := range listeners {
		_ = listener.Close()
		if listener.Addr().Network() == "unix" {
			_ = os.Remove(listener.Addr().String())
		}
	}
}

func removeUnixSocket(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSocket == 0 {
		return fmt.Errorf("refusing to remove non-socket file %q", path)
	}
	return os.Remove(path)
}
