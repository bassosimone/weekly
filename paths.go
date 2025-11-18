// paths.go - utilities to construct paths
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import "path/filepath"

// calendarPath returns the calendar.json path within dataDir.
func calendarPath(dataDir string) string {
	return filepath.Join(dataDir, "calendar.json")
}

// credentialsPath returns the credentials.json path within dataDir.
func credentialsPath(dataDir string) string {
	return filepath.Join(dataDir, "credentials.json")
}
