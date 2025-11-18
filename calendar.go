// calendar.go - utilities to manage calendar.json
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"encoding/json"

	"github.com/rogpeppe/go-internal/lockedfile"
)

// calendarInfo contains the selected calendar info.
type calendarInfo struct {
	// ID is the calendar unique identifier.
	ID string
}

// readCalendarInfo reads [*calendarInfo] from the given filePath.
func readCalendarInfo(path string) (*calendarInfo, error) {
	rawData, err := lockedfile.Read(path)
	if err != nil {
		return nil, err
	}
	var info calendarInfo
	if err := json.Unmarshal(rawData, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// writeCalendarInfo writes [*calendarInfo] to the given filePath.
func writeCalendarInfo(path string, info *calendarInfo) error {
	return lockedfile.Write(path, bytes.NewReader(must1(json.Marshal(info))), 0600)
}
