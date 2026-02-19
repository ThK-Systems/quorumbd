// Package config provides configuration loading and validation
package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

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
	Logging struct {
		Type     LoggingType `toml:"type"`
		FileName string      `toml:"filename"`
	} `toml:"logging"`
	Core struct {
		Server         string   `toml:"server"`
		FallbackServer []string `toml:"fallback_server"`
		DataPort       int      `toml:"data_port"`
		ControlPort    int      `toml:"control_port"`
	} `toml:"core"`
}

func Get() *Config {
	if config == nil {
		panic("config.Get() called before Load()")
	}
	return config
}

func Load() error {
	var initErr error

	once.Do(func() {
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

	if err := validateConfig(&cfg); err != nil {
		return fmt.Errorf("invalid config %s: %w", configPath, err)
	}

	config = &cfg

	return nil
}

func validateConfig(cfg *Config) error {
	err := validateConfigLogging(cfg)
	if err != nil {
		return err
	}
	// TODO: Validate common
	// TODO: Validate core
	return nil
}

func validateConfigLogging(cfg *Config) error {
	if cfg.Logging.Type == "" {
		cfg.Logging.Type = LoggingStdout
	}
	switch cfg.Logging.Type {
	case LoggingStdout:
		if cfg.Logging.FileName != "" {
			return fmt.Errorf("logging.filename must be empty when logging.type=%s", LoggingStdout)
		}
		return nil
	case LoggingFile:
		if cfg.Logging.FileName == "" {
			return fmt.Errorf("logging.filename required when logging.type=%s", LoggingFile)
		}
	default:
		return fmt.Errorf("invalid logging.type: %s", cfg.Logging.Type)
	}
	return nil
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
