// Package config provides common middleware configuration and validation
package config

import (
	"fmt"
	"net"
	"strings"

	commonconfig "thk-systems.net/quorumbd/common/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	CommonConfig         commonconfig.CommonConfig `toml:"common"`
	CoreConnectionConfig CoreConnectionConfig      `toml:"coreconnection"`
}

type CoreConnectionConfig struct {
	Server         string   `toml:"server"`
	ServerFallback []string `toml:"server_fallback"`
}

func (cfg *CoreConnectionConfig) SetDefaults() {
}

func (cfg *CoreConnectionConfig) Validate() error {
	return validation.Errors{
		"coreconnection": validation.ValidateStruct(cfg,
			validation.Field(&cfg.Server,
				validation.Required.Error("core.server required"),
				validation.By(validateCoreURI),
			),
			validation.Field(&cfg.ServerFallback,
				validation.Each(validation.By(validateCoreURI)),
			),
		)}.Filter()
}

func validateCoreURI(value interface{}) error {
	uri, ok := value.(string)
	if !ok {
		return fmt.Errorf("URI must be a string")
	}

	uri = strings.TrimSpace(uri)
	if uri == "" {
		return fmt.Errorf("URI must not be empty")
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
			return fmt.Errorf("invalid tcp URI %q: missing host:port", uri)
		}
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return fmt.Errorf("invalid tcp URI %q: %w", uri, err)
		}
		if net.ParseIP(host) == nil {
			return fmt.Errorf("invalid tcp URI %q: host must be a valid IP address", uri)
		}
		return nil

	default:
		return fmt.Errorf("invalid URI %q: expected unix:// or tcp://", uri)
	}
}
