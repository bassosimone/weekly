// must_test.go - tests for must.go
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"bytes"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMust0(t *testing.T) {
	// describes a test case run by this function
	type testCase struct {
		// name is the test case name
		name string

		// err is the error to pass to must0
		err error

		// expectExit is true if we expect the function to call Exit
		expectExit bool

		// expectExitCode is the expected exit code (if expectExit is true)
		expectExitCode int64

		// expectStderr is the expected stderr output (if expectExit is true)
		expectStderr string
	}

	// defines all test cases
	cases := []testCase{
		{
			name:           "nil error does not exit",
			err:            nil,
			expectExit:     false,
			expectExitCode: 0,
			expectStderr:   "",
		},

		{
			name:           "non-nil error exits with code 1",
			err:            errors.New("something went wrong"),
			expectExit:     true,
			expectExitCode: 1,
			expectStderr:   "fatal: something went wrong",
		},

		{
			name:           "wrapped error shows full message",
			err:            errors.New("file not found: /path/to/file"),
			expectExit:     true,
			expectExitCode: 1,
			expectStderr:   "fatal: file not found: /path/to/file",
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

			// capture stderr
			stderr := &bytes.Buffer{}
			env.OSStderr = stderr

			// capture exit call
			exitCalled := &atomic.Bool{}
			exitCode := &atomic.Int64{}
			errPanicSentinel := errors.New("exit called")
			env.OSExit = func(code int) {
				exitCalled.Store(true)
				exitCode.Store(int64(code))
				panic(errPanicSentinel)
			}

			// execute the function under test (with panic handling)
			func() {
				defer func() {
					if r := recover(); r != nil {
						err, ok := r.(error)
						if !ok || !errors.Is(err, errPanicSentinel) {
							t.Errorf("unexpected panic: %v", r)
						}
					}
				}()
				must0(env, tc.err)
			}()

			// check exit expectation
			if tc.expectExit && !exitCalled.Load() {
				t.Error("expected Exit to be called but it was not")
			}
			if !tc.expectExit && exitCalled.Load() {
				t.Error("expected Exit not to be called but it was")
			}

			// check exit code
			if tc.expectExit {
				if diff := cmp.Diff(tc.expectExitCode, exitCode.Load()); diff != "" {
					t.Error("exit code differs:", diff)
				}
			}

			// check stderr
			if tc.expectExit {
				if diff := cmp.Diff(tc.expectStderr, stderr.String()); diff != "" {
					t.Error("stderr differs:", diff)
				}
			}
		})
	}
}

func TestMust1(t *testing.T) {
	// describes a test case run by this function
	type testCase struct {
		// name is the test case name
		name string

		// value is the value to pass to must1
		value string

		// err is the error to pass to must1
		err error

		// expectExit is true if we expect the function to call Exit
		expectExit bool

		// expectValue is the expected return value (if not exiting)
		expectValue string
	}

	// defines all test cases
	cases := []testCase{
		{
			name:        "nil error returns value",
			value:       "success",
			err:         nil,
			expectExit:  false,
			expectValue: "success",
		},

		{
			name:        "non-nil error exits",
			value:       "ignored",
			err:         errors.New("failure"),
			expectExit:  true,
			expectValue: "",
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

			// capture stderr
			stderr := &bytes.Buffer{}
			env.OSStderr = stderr

			// capture exit call
			exitCalled := &atomic.Bool{}
			errPanicSentinel := errors.New("exit called")
			env.OSExit = func(code int) {
				exitCalled.Store(true)
				panic(errPanicSentinel)
			}

			// execute the function under test (with panic handling)
			var result string
			func() {
				defer func() {
					if r := recover(); r != nil {
						err, ok := r.(error)
						if !ok || !errors.Is(err, errPanicSentinel) {
							t.Errorf("unexpected panic: %v", r)
						}
					}
				}()
				result = must1(tc.value, tc.err)
			}()

			// check exit expectation
			if tc.expectExit && !exitCalled.Load() {
				t.Error("expected Exit to be called but it was not")
			}
			if !tc.expectExit && exitCalled.Load() {
				t.Error("expected Exit not to be called but it was")
			}

			// check return value
			if !tc.expectExit {
				if diff := cmp.Diff(tc.expectValue, result); diff != "" {
					t.Error("return value differs:", diff)
				}
			}
		})
	}
}
