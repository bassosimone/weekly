// xdg.go - xdg package implementation
// SPDX-License-Identifier: GPL-3.0-or-later

// Package xdg contains code to access XDG config variables.
package xdg

import (
	"errors"
	"path/filepath"
)

// ExecEnv abstracts [ConfigHome] dependencies.
type ExecEnv interface {
	// LookupEnv is equivalent to [os.LookupEnv].
	LookupEnv(key string) (string, bool)
}

// ConfigHome returns the directory containing the configuration.
func ConfigHome(env ExecEnv) (string, error) {
	if base, found := env.LookupEnv("XDG_CONFIG_HOME"); found {
		return filepath.Join(base, "weekly"), nil
	}
	if base, found := env.LookupEnv("HOME"); found {
		return filepath.Join(base, ".config", "weekly"), nil
	}
	return "", errors.New("neither $XDG_CONFIG_HOME nor $HOME is defined")
}
