// paths.go - utilities to construct paths
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"path/filepath"

	"github.com/bassosimone/weekly/internal/xdg"
)

// xdgConfigHome returns the directory containing config.
func xdgConfigHome(env xdg.ExecEnv) string {
	return must1(xdg.ConfigHome(env))
}

// calendarPath returns the calendar.json path within configDir.
func calendarPath(configDir string) string {
	return filepath.Join(configDir, "calendar.json")
}

// credentialsPath returns the credentials.json path within configDir.
func credentialsPath(configDir string) string {
	return filepath.Join(configDir, "credentials.json")
}
