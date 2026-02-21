// Package config provides common configuration loading and validation
package config

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type LoggingType string

const (
	LoggingStdout LoggingType = "stdout"
	LoggingFile   LoggingType = "file"
)

type LoggingConfig struct {
	Type     LoggingType `toml:"type"`
	FileName string      `toml:"filename"`
	LogLevel string      `toml:"log_level"`
}

func SetDefaults(cfg *LoggingConfig) {
	// Logging
	if cfg.Type == "" {
		cfg.Type = LoggingStdout
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "INFO"
	}
}

func ValidateConfig(cfg LoggingConfig) error {
	return validation.Errors{
		"logging": validation.ValidateStruct(&cfg,
			validation.Field(&cfg.Type, validation.Required.Error("logging.type required"), validation.In(LoggingStdout, LoggingFile).Error("invalid logging.type")),
			validation.Field(&cfg.FileName, validation.When(cfg.Type == LoggingFile, validation.Required.Error("logging.filename required when logging.type=file")),
				validation.When(cfg.Type == LoggingStdout, validation.Empty.Error("logging.filename must be empty when logging.type=stdout"))),
			validation.Field(&cfg.LogLevel, validation.Required.Error("logging.log_level required"), validation.In(slog.LevelDebug.String(), slog.LevelInfo.String(), slog.LevelWarn.String(), slog.LevelError.String()).Error("invalid logging.log_level"))),
	}.Filter()
}

func ResolveConfigPath(cfgFileName string) (string, error) {
	// 1. ENV
	if env := os.Getenv("QUORUMBD_NBDSERVER_CONFIG"); env != "" {
		if fileExists(env) {
			return env, nil
		}
	}

	// 2. User config (~/.config/quorumbd/<config-file>)
	if dir, err := os.UserConfigDir(); err == nil {
		path := filepath.Join(dir, "quorumbd", cfgFileName)
		if fileExists(path) {
			return path, nil
		}
	}

	// 3. User home (~/.quorumbd/<config-file>)
	if dir, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(dir, ".quorumbd", cfgFileName)
		if fileExists(path) {
			return path, nil
		}
	}

	// 4. System config
	path := filepath.Join("/etc/", "quorumbd", cfgFileName)
	if fileExists(path) {
		return path, nil
	}

	// No config
	return "", errors.New("no config file found")
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func MergeValidationErrors(errs ...error) error {
	merged := validation.Errors{}

	for _, err := range errs {
		if err == nil {
			continue
		}

		var ve validation.Errors
		if errors.As(err, &ve) {
			for k, v := range ve {
				merged[k] = v
			}
		} else {
			return err
		}
	}

	if len(merged) == 0 {
		return nil
	}
	return merged
}
