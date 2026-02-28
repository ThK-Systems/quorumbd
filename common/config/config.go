// Package config provides common configuration and validation
package config

import (
	"errors"
	"log/slog"
	"maps"
	"os"
	"path/filepath"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	LoggingType   string
	LoggingFormat string
)

const (
	LoggingTypeStdout LoggingType = "stdout"
	LoggingTypeFile   LoggingType = "file"
)

const (
	LoggingFormatText LoggingFormat = "text"
	LoggingFormatJSON LoggingFormat = "json"
)

type CommonConfig struct {
	StateDir string `toml:"state_dir"`
}

type LoggingConfig struct {
	Type     LoggingType   `toml:"type"`
	FileName string        `toml:"filename"`
	Level    string        `toml:"level"`
	Format   LoggingFormat `toml:"format"`
}

func (cfg *LoggingConfig) SetDefaults() {
	cfg.Type = LoggingTypeStdout
	cfg.Level = "INFO"
	cfg.Format = "text"
}

func (cfg *CommonConfig) SetDefaults() {
	cfg.StateDir = filepath.Join("/", "var", "lib", "state", "quorumbd")
}

func (cfg *LoggingConfig) Validate() error {
	return validation.Errors{
		"logging": validation.ValidateStruct(cfg,
			validation.Field(&cfg.Type, validation.Required.Error("logging.type required"), validation.In(LoggingTypeStdout, LoggingTypeFile).Error("invalid logging.type")),
			validation.Field(&cfg.Format, validation.Required.Error("logging.format required"), validation.In(LoggingFormatJSON, LoggingFormatText).Error("invalid logging.format")),
			validation.Field(&cfg.FileName, validation.When(cfg.Type == LoggingTypeFile, validation.Required.Error("logging.filename required when logging.type=file")),
				validation.When(cfg.Type == LoggingTypeStdout, validation.Empty.Error("logging.filename must be empty when logging.type=stdout"))),
			validation.Field(&cfg.Level, validation.Required.Error("logging.level required"), validation.In(slog.LevelDebug.String(), slog.LevelInfo.String(), slog.LevelWarn.String(), slog.LevelError.String()).Error("invalid logging.log_level"))),
	}.Filter()
}

func (cfg *CommonConfig) Validate() error {
	return validation.Errors{
		"common": validation.ValidateStruct(cfg, validation.Field(&cfg.StateDir, validation.Required.Error("common.state_dir required"))),
	}.Filter()
}

func ResolveConfigPath(cfgFileName string, envVarName string) (string, error) {
	// 1. ENV
	if envVarName != "" {
		if env := os.Getenv(envVarName); env != "" {
			if fileExists(env) {
				return env, nil
			}
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
		if ve, ok := errors.AsType[validation.Errors](err); ok {
			maps.Copy(merged, ve)
		} else {
			return err
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}
