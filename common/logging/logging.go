// Package logging provides logging functionality
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"thk-systems.net/quorumbd/common/config"
)

func Init(cfg config.LoggingConfig) error {
	var writer io.Writer
	var err error

	switch cfg.Type {
	case config.LoggingStdout:
		writer = os.Stdout

	case config.LoggingFile:
		writer, err = os.OpenFile(cfg.FileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("cannot open logging file: %w", err)
		}

	default:
		return fmt.Errorf("unknown logging type: %s", cfg.Type)
	}

	// Set default log level
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", cfg.LogLevel, err)
	}

	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: &level,
	})
	slog.SetDefault(slog.New(handler))

	return nil
}

func For(pkg string) *slog.Logger {
	return slog.Default().With(slog.String("pkg", pkg))
}
