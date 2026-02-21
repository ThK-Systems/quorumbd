// Package logging provides logging functionality
package logging

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"thk-systems.net/quorumbd/common/config"
)

func Init(cfg config.LoggingConfig) error {
	var writer *os.File
	var err error

	switch cfg.Type {
	case config.LoggingStdout:
		writer = os.Stdout

	case config.LoggingFile:
		writer, err = os.OpenFile(
			cfg.FileName,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY,
			0644,
		)
		if err != nil {
			return fmt.Errorf("cannot open logging file: %w", err)
		}

	default:
		return fmt.Errorf("unknown logging type: %s", cfg.Type)
	}

	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String(slog.TimeKey, a.Value.Time().Format(time.RFC3339))
			}
			return a
		},
	})

	slog.SetDefault(slog.New(handler))
	return nil
}
