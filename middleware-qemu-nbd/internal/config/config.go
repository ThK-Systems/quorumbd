// Package config provides configuration loading and validation
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	commonconfig "thk-systems.net/quorumbd/common/config"
	middelwareconfig "thk-systems.net/quorumbd/middleware-common/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	toml "github.com/pelletier/go-toml/v2"
)

const configFileName = "middleware-qemu-nbd.toml"

var (
	config  *Config
	once    sync.Once
	loadErr error
)

type nbdServer struct {
	Socket string `toml:"socket"`
}

type Config struct {
	Common         commonconfig.CommonConfig       `toml:"common"`
	Logging        commonconfig.LoggingConfig      `toml:"logging"`
	CoreConnection middelwareconfig.CoreConnection `toml:"coreconnection"`
	NBDServer      nbdServer                       `toml:"nbdserver"`
}

func Get() *Config {
	if config == nil {
		panic("config.Get() called before Load()") // This can never happen
	}
	return config
}

func (cfg *Config) ToMiddlewareConfig() *middelwareconfig.Config {
	return &middelwareconfig.Config{
		Common:         cfg.Common,
		CoreConnection: cfg.CoreConnection,
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
	cfg.Common.SetDefaults()
	cfg.Logging.SetDefaults()
	cfg.CoreConnection.SetDefaults()
	cfg.NBDServer.setDefaults()
}

func (cfg *nbdServer) setDefaults() {
	cfg.Socket = filepath.Join("/", "var", "run", "quorumbd", "main.sock")
}

func (cfg *Config) validate() error {
	commonErrors := cfg.Common.Validate()
	loggingErrors := cfg.Logging.Validate()
	coreConnectionErrors := cfg.CoreConnection.Validate()
	nbdServerErrors := cfg.NBDServer.validate()
	return commonconfig.MergeValidationErrors(commonErrors, loggingErrors, coreConnectionErrors, nbdServerErrors)
}

func (cfg *nbdServer) validate() error {
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
