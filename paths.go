// paths.go - utilities to construct paths
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"path/filepath"
)

// xdgConfigHome returns the directory containing config.
func xdgConfigHome() string {
	base, found := os.LookupEnv("XDG_CONFIG_HOME")
	if !found {
		base = filepath.Join(os.ExpandEnv("${HOME}"), ".config")
	}
	return filepath.Join(base, "weekly")
}

// calendarPath returns the calendar.json path within configDir.
func calendarPath(configDir string) string {
	return filepath.Join(configDir, "calendar.json")
}

// credentialsPath returns the credentials.json path within configDir.
func credentialsPath(configDir string) string {
	return filepath.Join(configDir, "credentials.json")
}
