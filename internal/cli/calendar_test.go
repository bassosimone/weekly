// calendar_test.go - tests for calendar.go
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"errors"
	"io"
	"io/fs"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadCalendarInfo(t *testing.T) {
	// describes a test case run by this function
	type testCase struct {
		// name is the test case name
		name string

		// lockedfileRead is the mock function for reading files
		lockedfileRead func(path string) ([]byte, error)

		// path is the path to read from
		path string

		// expectInfo is the expected calendar info (nil if expecting error)
		expectInfo *calendarInfo

		// expectError is true if we expect an error
		expectError bool
	}

	// defines all test cases
	cases := []testCase{
		{
			name: "successful read with valid JSON",
			lockedfileRead: func(path string) ([]byte, error) {
				return []byte(`{"ID":"test-calendar-id"}`), nil
			},
			path: "/path/to/calendar.json",
			expectInfo: &calendarInfo{
				ID: "test-calendar-id",
			},
			expectError: false,
		},

		{
			name: "file read error",
			lockedfileRead: func(path string) ([]byte, error) {
				return nil, fs.ErrNotExist
			},
			path:        "/nonexistent/calendar.json",
			expectInfo:  nil,
			expectError: true,
		},

		{
			name: "invalid JSON",
			lockedfileRead: func(path string) ([]byte, error) {
				return []byte(`{this is not valid json}`), nil
			},
			path:        "/path/to/calendar.json",
			expectInfo:  nil,
			expectError: true,
		},

		{
			name: "empty JSON object",
			lockedfileRead: func(path string) ([]byte, error) {
				return []byte(`{}`), nil
			},
			path: "/path/to/calendar.json",
			expectInfo: &calendarInfo{
				ID: "",
			},
			expectError: false,
		},

		{
			name: "malformed JSON - truncated",
			lockedfileRead: func(path string) ([]byte, error) {
				return []byte(`{"ID":"test`), nil
			},
			path:        "/path/to/calendar.json",
			expectInfo:  nil,
			expectError: true,
		},
	}

	// run each test case
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Save and restore the global env
			oldEnv := env
			defer func() {
				env = oldEnv
			}()

			// create test environment
			env = newExecEnv()
			env.LockedfileRead = tc.lockedfileRead

			// execute the function under test
			info, err := readCalendarInfo(env, tc.path)

			// check error expectation
			if tc.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// check info matches expectation
			if diff := cmp.Diff(tc.expectInfo, info); diff != "" {
				t.Error("calendar info differs:", diff)
			}
		})
	}
}

func TestWriteCalendarInfo(t *testing.T) {
	// describes a test case run by this function
	type testCase struct {
		// name is the test case name
		name string

		// lockedfileWrite is the mock function for writing files
		lockedfileWrite func(path string, content io.Reader, perms fs.FileMode) error

		// path is the path to write to
		path string

		// info is the calendar info to write
		info *calendarInfo

		// expectError is true if we expect an error
		expectError bool

		// expectWrittenData is the data we expect to be written (if success)
		expectWrittenData string
	}

	// defines all test cases
	cases := []testCase{
		{
			name: "successful write",
			lockedfileWrite: func(path string, content io.Reader, perms fs.FileMode) error {
				// read and verify the content
				data, err := io.ReadAll(content)
				if err != nil {
					return err
				}
				expectedData := `{"ID":"test-calendar-id"}`
				if string(data) != expectedData {
					t.Errorf("expected data %q but got %q", expectedData, string(data))
				}
				if perms != 0600 {
					t.Errorf("expected perms 0600 but got %o", perms)
				}
				return nil
			},
			path: "/path/to/calendar.json",
			info: &calendarInfo{
				ID: "test-calendar-id",
			},
			expectError:       false,
			expectWrittenData: `{"ID":"test-calendar-id"}`,
		},

		{
			name: "write error - permission denied",
			lockedfileWrite: func(path string, content io.Reader, perms fs.FileMode) error {
				return fs.ErrPermission
			},
			path: "/readonly/calendar.json",
			info: &calendarInfo{
				ID: "test-calendar-id",
			},
			expectError: true,
		},

		{
			name: "write error - generic error",
			lockedfileWrite: func(path string, content io.Reader, perms fs.FileMode) error {
				return errors.New("disk full")
			},
			path: "/path/to/calendar.json",
			info: &calendarInfo{
				ID: "test-calendar-id",
			},
			expectError: true,
		},

		{
			name: "write empty ID",
			lockedfileWrite: func(path string, content io.Reader, perms fs.FileMode) error {
				data, err := io.ReadAll(content)
				if err != nil {
					return err
				}
				expectedData := `{"ID":""}`
				if string(data) != expectedData {
					t.Errorf("expected data %q but got %q", expectedData, string(data))
				}
				return nil
			},
			path: "/path/to/calendar.json",
			info: &calendarInfo{
				ID: "",
			},
			expectError:       false,
			expectWrittenData: `{"ID":""}`,
		},
	}

	// run each test case
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Save and restore the global env
			oldEnv := env
			defer func() {
				env = oldEnv
			}()

			// create test environment
			env = newExecEnv()
			env.LockedfileWrite = tc.lockedfileWrite

			// execute the function under test
			err := writeCalendarInfo(env, tc.path, tc.info)

			// check error expectation
			if tc.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
