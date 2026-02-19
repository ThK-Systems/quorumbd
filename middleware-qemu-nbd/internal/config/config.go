// Package config provides configuration loading and validation
package config

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pelletier/go-toml/v2"
)

const configFileName = "middleware-qemu-nbd.toml"

type LoggingType string

const (
	LoggingStdout LoggingType = "stdout"
	LoggingFile   LoggingType = "file"
)

var (
	config *Config
	once   sync.Once
)

type Config struct {
	Common struct {
		StateDir string `toml:"state_dir"`
	} `toml:"common"`
	NBDServer struct {
		Socket string `toml:"socket"`
	} `toml:"nbdserver"`
	Core struct {
		Server          string   `toml:"server"`
		ServerFallback  []string `toml:"server_fallback"`
		Control         string   `toml:"control"`
		ControlFallback []string `toml:"control_fallback"`
	} `toml:"core"`
	Logging struct {
		Type     LoggingType `toml:"type"`
		FileName string      `toml:"filename"`
		LogLevel string      `toml:"log_level"`
	} `toml:"logging"`
}

func Get() *Config {
	if config == nil {
		panic("config.Get() called before Load()")
	}
	return config
}

func Load() error {
	var initErr error

	once.Do(func() { // => Singleton
		if err := load(); err != nil {
			initErr = err
			return
		}
	})

	return initErr
}

func load() error {
	configPath, err := resolveConfigPath()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	var cfg Config

	if err := readConfig(configPath, &cfg); err != nil {
		return fmt.Errorf("parsing config %s: %w", configPath, err)
	}

	setDefaults(&cfg)

	if err := validateConfig(&cfg); err != nil {
		return fmt.Errorf("invalid config %s: %w", configPath, err)
	}

	config = &cfg

	return nil
}

func setDefaults(cfg *Config) {
	// Common
	if cfg.Common.StateDir == "" {
		cfg.Common.StateDir = filepath.Join("var", "lib", "state", "quorumbd", "middleware-qemu-nbd")
	}

	// Logging
	if cfg.Logging.Type == "" {
		cfg.Logging.Type = LoggingStdout
	}
	if cfg.Logging.LogLevel == "" {
		cfg.Logging.LogLevel = "INFO"
	}
}

func validateConfig(cfg *Config) error {
	return validation.Errors{
		"common": validation.ValidateStruct(&cfg.Common, validation.Field(&cfg.Common.StateDir, validation.Required.Error("common.state_dir required"))),

		"nbdserver": validation.ValidateStruct(&cfg.NBDServer, validation.Field(&cfg.NBDServer.Socket, validation.Required.Error("nbdserver.main_socket required"))),

		"core": validation.ValidateStruct(&cfg.Core,
			validation.Field(&cfg.Core.Server, validation.Required.Error("core.server required")),
			validation.Field(&cfg.Core.Control, validation.Required.Error("core.control required")),
			validation.Field(&cfg.Core.ControlFallback, validation.When(cfg.Core.ServerFallback != nil, validation.Required.Error("core.control_fallback is required when core.server_fallback is set"))),
			validation.Field(&cfg.Core.ControlFallback, validation.When(cfg.Core.ControlFallback != nil, validation.Length(len(cfg.Core.ServerFallback), len(cfg.Core.ServerFallback)).Error("core.control_fallback must be of same length as core.server_fallback"))),
		),

		"logging": validation.ValidateStruct(&cfg.Logging,
			validation.Field(&cfg.Logging.Type, validation.Required.Error("logging.type required"), validation.In(LoggingStdout, LoggingFile).Error("invalid logging.type")),
			validation.Field(&cfg.Logging.FileName, validation.When(cfg.Logging.Type == LoggingFile, validation.Required.Error("logging.filename required when logging.type=file")),
				validation.When(cfg.Logging.Type == LoggingStdout, validation.Empty.Error("logging.filename must be empty when logging.type=stdout"))),
			validation.Field(&cfg.Logging.LogLevel, validation.Required.Error("logging.log_level required"), validation.In(slog.LevelDebug.String(), slog.LevelInfo.String(), slog.LevelWarn.String(), slog.LevelError.String()).Error("invalid logging.log_level"))),
	}.Filter()
}

func readConfig(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	decoder := toml.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(cfg); err != nil {
		return err
	}

	return nil
}

func resolveConfigPath() (string, error) {
	// 1. ENV
	if env := os.Getenv("QUORUMBD_NBDSERVER_CONFIG"); env != "" {
		if fileExists(env) {
			return env, nil
		}
	}

	// 2. User config (~/.config/...)
	if dir, err := os.UserConfigDir(); err == nil {
		path := filepath.Join(dir, "quorumbd", configFileName)
		if fileExists(path) {
			return path, nil
		}
	}

	// 3. System config
	path := filepath.Join("/etc/", "quorumbd", configFileName)
	if fileExists(path) {
		return path, nil
	}

	return "", errors.New("no config file found")
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
