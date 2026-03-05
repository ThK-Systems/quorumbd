// Package config provides common middleware configuration and validation
package config

import (
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
			validation.Field(&cfg.Server, validation.Required.Error("core.server required")),
		)}.Filter()
}
