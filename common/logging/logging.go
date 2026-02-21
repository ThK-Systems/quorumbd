// Package logging provides logging functionality
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"thk-systems.net/quorumbd/common/config"
)

var once sync.Once

func Initialize(cfg config.LoggingConfig) error {
	var initErr error
	once.Do(func() { // => Singleton
		if err := initialize(cfg); err != nil {
			initErr = err
			return
		}
	})
	return initErr
}

func initialize(cfg config.LoggingConfig) error {
	var writer io.Writer
	var err error

	switch cfg.Type {
	case config.LoggingTypeStdout:
		writer = os.Stdout

	case config.LoggingTypeFile:
		writer, err = os.OpenFile(cfg.FileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("cannot open logging file: %w", err)
		}
	}

	// Set default log level
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", cfg.Level, err)
	}

	var handler slog.Handler
	switch cfg.Format {
	case config.LoggingFormatJSON:
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})
	case config.LoggingFormatText:
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(handler))

	return nil
}

func For(pkg string) *slog.Logger {
	return slog.Default().With(slog.String("pkg", pkg))
}
