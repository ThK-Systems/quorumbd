// Main package of core
package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	startFakeCore()
}

func startFakeCore() {
	var (
		tcpAddr  = flag.String("tcp", "127.0.0.1:7447", "TCP listen address (empty disables)")
		unixPath = flag.String("unix", "", "Unix socket path (empty disables)")
		hold     = flag.Duration("hold", 0, "Hold accepted connections for this duration before closing (0 = close immediately)")
	)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if *tcpAddr == "" && *unixPath == "" {
		log.Fatal("either -tcp or -unix must be set")
	}

	errCh := make(chan error, 2)

	if *tcpAddr != "" {
		ln, err := net.Listen("tcp", *tcpAddr)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("FakeCore listening on tcp://%s", *tcpAddr)
		go func() { errCh <- serve(ctx, ln, *hold) }()
	}

	if *unixPath != "" {
		_ = os.Remove(*unixPath)
		ln, err := net.Listen("unix", *unixPath)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("FakeCore listening on unix://%s", *unixPath)
		go func() { errCh <- serve(ctx, ln, *hold) }()
	}

	select {
	case <-ctx.Done():
		// graceful shutdown
		log.Printf("FakeCore shutting down...")
		return
	case err := <-errCh:
		log.Printf("FakeCore error: %v", err)
		os.Exit(1)
	}
}

func serve(ctx context.Context, ln net.Listener, hold time.Duration) error {
	defer ln.Close()

	for {
		// Accept blocks; closing ln on ctx cancel will unblock on most platforms.
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return err
			}
		}

		go func(c net.Conn) {
			defer c.Close()
			if hold > 0 {
				time.Sleep(hold)
			}
		}(conn)
	}
}
