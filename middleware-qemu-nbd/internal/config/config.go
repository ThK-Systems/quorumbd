// Package config provides configuration loading and validation
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	commonconfig "quorumbd.net/common/config"
	middlewareconfig "quorumbd.net/middleware-common/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	toml "github.com/pelletier/go-toml/v2"
)

const configFileName = "middleware-qemu-nbd.toml"

var (
	config  *Config
	once    sync.Once
	loadErr error
)

type nbdServerConfig struct {
	Socket string `toml:"socket"`
}

type Config struct {
	CommonConfig         commonconfig.CommonConfig             `toml:"common"`
	LoggingConfig        commonconfig.LoggingConfig            `toml:"logging"`
	CoreConnectionConfig middlewareconfig.CoreConnectionConfig `toml:"coreconnection"`
	NBDServerConfig      nbdServerConfig                       `toml:"nbdserver"`
}

func Get() *Config {
	if config == nil {
		panic("config.Get() called before Load()") // This can never happen
	}
	return config
}

func (cfg *Config) ToMiddlewareConfig() *middlewareconfig.Config {
	return &middlewareconfig.Config{
		CommonConfig:         cfg.CommonConfig,
		CoreConnectionConfig: cfg.CoreConnectionConfig,
	}
}

func Load() error {
	once.Do(func() { // => Singleton
		loadErr = load()
	})
	return loadErr
}

func load() error {
	configPath, err := commonconfig.ResolveConfigPath(configFileName, "QUORUMBD_NBDSERVER_CONFIG")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	var cfg Config

	cfg.setDefaults()

	if err := cfg.readConfig(configPath); err != nil {
		return fmt.Errorf("parsing config %s: %w", configPath, err)
	}

	if err := cfg.validate(); err != nil {
		return fmt.Errorf("invalid config %s: %w", configPath, err)
	}

	config = &cfg

	return nil
}

func (cfg *Config) setDefaults() {
	cfg.CommonConfig.SetDefaults()
	cfg.LoggingConfig.SetDefaults()
	cfg.CoreConnectionConfig.SetDefaults()
	cfg.NBDServerConfig.setDefaults()
}

func (cfg *nbdServerConfig) setDefaults() {
	cfg.Socket = filepath.Join("/", "var", "run", "quorumbd", "main.sock")
}

func (cfg *Config) validate() error {
	commonErrors := cfg.CommonConfig.Validate()
	loggingErrors := cfg.LoggingConfig.Validate()
	coreConnectionErrors := cfg.CoreConnectionConfig.Validate()
	nbdServerErrors := cfg.NBDServerConfig.validate()
	return commonconfig.MergeValidationErrors(commonErrors, loggingErrors, coreConnectionErrors, nbdServerErrors)
}

func (cfg *nbdServerConfig) validate() error {
	return validation.Errors{
		"nbdserver": validation.ValidateStruct(cfg, validation.Field(&cfg.Socket, validation.Required.Error("nbdserver.socket required"))),
	}.Filter()
}

func (cfg *Config) readConfig(path string) error {
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
