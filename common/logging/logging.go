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

var (
	once    sync.Once
	initErr error
	writer  io.Writer
	file    *os.File
)

func Initialize(cfg config.LoggingConfig) error {
	once.Do(func() { // => Singleton
		initErr = initialize(cfg)
	})
	return initErr
}

func initialize(cfg config.LoggingConfig) error {
	switch cfg.Type {
	case config.LoggingTypeStdout:
		writer = os.Stdout

	case config.LoggingTypeFile:
		file, err := os.OpenFile(cfg.FileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("cannot open logging file: %w", err)
		}
		writer = file
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

func TerminateLogging(cfg config.LoggingConfig) error {
	if file != nil {
		return file.Close()
	}
	return nil
}

func GetDefaultLogger() *slog.Logger {
	return slog.Default()
}

func With(args ...any) *slog.Logger {
	return GetDefaultLogger().With(args...)
}

func WithSingle(key string, value string) *slog.Logger {
	return With(key, value)
}

func For(pkg string) *slog.Logger {
	return WithSingle("pkg", pkg)
}
