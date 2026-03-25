// Package config provides configuration loading and validation
package config

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pelletier/go-toml/v2"
	commonconfig "quorumbd.net/common/config"
)

const configFileName = "core.toml"

var (
	config  *Config
	once    sync.Once
	loadErr error
)

type Config struct {
	CommonConfig  commonconfig.CommonConfig  `toml:"common"`
	LoggingConfig commonconfig.LoggingConfig `toml:"logging"`
	CoreConfig    coreConfig                 `toml:"core"`
}

type coreConfig struct {
	Listen []string `toml:"listen"`
}

func Load() (*Config, error) {
	once.Do(func() { // => Singleton
		config, loadErr = load()
	})
	return config, loadErr
}

func load() (*Config, error) {
	configPath, err := commonconfig.ResolveConfigPath(configFileName, "QUORUMBD_NBDSERVER_CONFIG")
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	var cfg Config

	cfg.setDefaults()

	if err := cfg.readConfig(configPath); err != nil {
		return nil, fmt.Errorf("error parsing config %s: %w", configPath, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("error invalid config %s: %w", configPath, err)
	}

	config = &cfg

	return config, nil
}

func (cfg *Config) setDefaults() {
	cfg.CommonConfig.SetDefaults()
	cfg.LoggingConfig.SetDefaults()
	cfg.CoreConfig.setDefaults()
}

func (cfg *coreConfig) setDefaults() {
	cfg.Listen = []string{"unix://" + filepath.Join("/", "var", "run", "qbd", "core.sock")}
}

func (cfg *Config) validate() error {
	commonErrors := cfg.CommonConfig.Validate()
	loggingErrors := cfg.LoggingConfig.Validate()
	coreErrors := cfg.CoreConfig.validate()
	return commonconfig.MergeValidationErrors(commonErrors, loggingErrors, coreErrors)
}

func (cfg coreConfig) validate() error {
	return validation.Errors{
		"core": validation.ValidateStruct(&cfg,
			validation.Field(&cfg.Listen,
				validation.Required.Error("core.listen required"),
				validation.By(func(value interface{}) error {
					listen, ok := value.([]string)
					if !ok {
						return fmt.Errorf("core.listen must be a list of addresses")
					}
					if len(listen) == 0 {
						return fmt.Errorf("core.listen must contain at least one listen address")
					}
					return nil
				}),
				validation.Each(validation.By(func(value interface{}) error {
					uri, ok := value.(string)
					if !ok {
						return fmt.Errorf("listen address must be a string")
					}

					uri = strings.TrimSpace(uri)
					if uri == "" {
						return fmt.Errorf("listen address must not be empty")
					}

					switch {
					case strings.HasPrefix(uri, "unix://"):
						address := strings.TrimSpace(strings.TrimPrefix(uri, "unix://"))
						if address == "" {
							return fmt.Errorf("invalid unix URI %q: missing socket path", uri)
						}
						return nil

					case strings.HasPrefix(uri, "tcp://"):
						address := strings.TrimSpace(strings.TrimPrefix(uri, "tcp://"))
						if address == "" {
							return fmt.Errorf("invalid tcp URI %q: missing host or host:port", uri)
						}
						if strings.Contains(address, ":") {
							host, _, err := net.SplitHostPort(address)
							if err != nil {
								return fmt.Errorf("invalid tcp URI %q: %w", uri, err)
							}
							if net.ParseIP(host) == nil {
								return fmt.Errorf("invalid tcp URI %q: host must be a valid IP address", uri)
							}
							return nil
						}
						if net.ParseIP(address) == nil {
							return fmt.Errorf("invalid tcp URI %q: host must be a valid IP address", uri)
						}
						return nil

					default:
						return fmt.Errorf("invalid URI %q: expected unix:// or tcp://", uri)
					}
				})),
			),
		),
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
