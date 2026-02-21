// Package config provides configuration loading and validation
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pelletier/go-toml/v2"
	commonconfig "thk-systems.net/quorumbd/common/config"
)

const configFileName = "middleware-qemu-nbd.toml"

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
	Logging commonconfig.LoggingConfig `toml:"logging"`
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
	configPath, err := commonconfig.ResolveConfigPath(configFileName)
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
	if cfg.Common.StateDir == "" {
		cfg.Common.StateDir = filepath.Join("var", "lib", "state", "quorumbd", "middleware-qemu-nbd")
	}
	commonconfig.SetLoggingDefaults(&cfg.Logging)
}

func validateConfig(cfg *Config) error {
	commonErrors := commonconfig.ValidateLoggingConfig(cfg.Logging)
	localErrors := validation.Errors{
		"common": validation.ValidateStruct(&cfg.Common, validation.Field(&cfg.Common.StateDir, validation.Required.Error("common.state_dir required"))),

		"nbdserver": validation.ValidateStruct(&cfg.NBDServer, validation.Field(&cfg.NBDServer.Socket, validation.Required.Error("nbdserver.main_socket required"))),

		"core": validation.ValidateStruct(&cfg.Core,
			validation.Field(&cfg.Core.Server, validation.Required.Error("core.server required")),
			validation.Field(&cfg.Core.Control, validation.Required.Error("core.control required")),
			validation.Field(&cfg.Core.ControlFallback, validation.When(cfg.Core.ServerFallback != nil, validation.Required.Error("core.control_fallback is required when core.server_fallback is set"))),
			validation.Field(&cfg.Core.ControlFallback, validation.When(cfg.Core.ControlFallback != nil, validation.Length(len(cfg.Core.ServerFallback), len(cfg.Core.ServerFallback)).Error("core.control_fallback must be of same length as core.server_fallback"))),
		),
	}.Filter()
	return commonconfig.MergeValidationErrors(commonErrors, localErrors)
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
