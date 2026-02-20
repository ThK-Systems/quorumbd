// Package config provides common configuration loading and validation
package config

import (
	"errors"
	"os"
	"path/filepath"
)

func ResolveConfigPath(cfgFileName string) (string, error) {
	// 1. ENV
	if env := os.Getenv("QUORUMBD_NBDSERVER_CONFIG"); env != "" {
		if fileExists(env) {
			return env, nil
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
