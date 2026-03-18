// Package errorhelper provides helper functions for working with errors
package errorhelper

import (
	"context"
	"errors"
	"io"
	"net"
	"syscall"
)

type ExitKind int

const (
	ExitShutdown ExitKind = iota
	ExitReconnect
	ExitFatal
)

func (k ExitKind) String() string {
	switch k {
	case ExitShutdown:
		return "shutdown"
	case ExitReconnect:
		return "reconnect"
	case ExitFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

type FatalConnError struct {
	error
}

func Fatal(err error) error {
	if err == nil {
		return nil
	}
	return &FatalConnError{err}
}

func ClassifyError(err error) ExitKind {
	if err == nil {
		return ExitShutdown
	}

	// explicit fatal (highest priority)
    if _, ok := errors.AsType[*FatalConnError](err); ok {
        return ExitFatal
    }

	// shutdown
	if errors.Is(err, context.Canceled) {
		return ExitShutdown
	}

	// EOF / closed
	if errors.Is(err, io.EOF) ||
		errors.Is(err, net.ErrClosed) {
		return ExitReconnect
	}

	// syscall-level
	if errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EPIPE) {
		return ExitReconnect
	}

	// net errors (timeouts etc.)
	if _, ok := errors.AsType[net.Error](err); ok {
		return ExitReconnect
	}

	// op error (very common wrapper)
	if _, ok := errors.AsType[*net.OpError](err); ok {
		return ExitReconnect
	}

	// fallback
	return ExitFatal
}
