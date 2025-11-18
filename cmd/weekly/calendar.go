// calendar.go - utilities to manage calendar.json
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"encoding/json"
	"os"
)

// calendarInfo contains the selected calendar info.
type calendarInfo struct {
	// ID is the calendar unique identifier.
	ID string
}

// readCalendarInfo reads [*calendarInfo] from the given filePath.
func readCalendarInfo(path string) (*calendarInfo, error) {
	rawData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var info calendarInfo
	if err := json.Unmarshal(rawData, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
