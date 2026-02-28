// Package config provides common middleware configuration and validation
package config

import (
	commonconfig "thk-systems.net/quorumbd/common/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Common         commonconfig.CommonConfig `toml:"common"`
	CoreConnection CoreConnection            `toml:"coreconnection"`
}

type CoreConnection struct {
	Server         string   `toml:"server"`
	ServerFallback []string `toml:"server_fallback"`
}

func (cfg *CoreConnection) SetDefaults() {
}

func (cfg *CoreConnection) Validate() error {
	return validation.Errors{
		"coreConnection": validation.ValidateStruct(cfg,
			validation.Field(&cfg.Server, validation.Required.Error("core.server required")),
		)}.Filter()
}
