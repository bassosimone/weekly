// xdg_test.go - xdg package tests
// SPDX-License-Identifier: GPL-3.0-or-later

package xdg

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// lookupEnv implements [ExecEnv]
type lookupEnv func(key string) (string, bool)

var _ ExecEnv = lookupEnv(nil)

// LookupEnv implements [ExecEnv].
func (fx lookupEnv) LookupEnv(key string) (string, bool) {
	return fx(key)
}

func TestConfigHome(t *testing.T) {

	// type describing test cases implemented by this function
	type testCase struct {
		// name is the name of the test case
		name string

		// lookupEnv is the function to mock env lookups
		lookupEnv func(key string) (string, bool)

		// output is the expected output
		output string

		// err is the expected error
		err error
	}

	// define all the test cases
	cases := []testCase{
		{
			name: "with no variable being set",
			lookupEnv: func(key string) (string, bool) {
				return "", false
			},
			output: "",
			err:    errors.New("neither $XDG_CONFIG_HOME nor $HOME is defined"),
		},

		{
			name: "with XDG_CONFIG_HOME being set",
			lookupEnv: func(key string) (string, bool) {
				switch key {
				case "XDG_CONFIG_HOME":
					return "foo", true
				default:
					return "", false
				}
			},
			output: filepath.Join("foo", "weekly"),
			err:    nil,
		},

		{
			name: "with HOME being set",
			lookupEnv: func(key string) (string, bool) {
				switch key {
				case "HOME":
					return "bar", true
				default:
					return "", false
				}
			},
			output: filepath.Join("bar", ".config", "weekly"),
			err:    nil,
		},

		{
			name: "with both XDG_CONFIG_HOME and HOME being set",
			lookupEnv: func(key string) (string, bool) {
				switch key {
				case "XDG_CONFIG_HOME":
					return "foo", true
				case "HOME":
					return "bar", true
				default:
					return "", false
				}
			},
			output: filepath.Join("foo", "weekly"),
			err:    nil,
		},
	}

	// run the test cases
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// invoke the function we're testing
			output, err := ConfigHome(lookupEnv(tc.lookupEnv))

			// make sure the error is the one we actually expect
			switch {
			case err == nil && tc.err == nil:
				// nothing

			case err == nil && tc.err != nil:
				t.Error("expected", tc.err, "got", err)

			case err != nil && tc.err == nil:
				t.Error("expected", tc.err, "got", err)

			case err != nil && tc.err != nil && err.Error() == tc.err.Error():
				// nothing

			default:
				t.Error("expected", tc.err, "got", err)
			}

			// make sure the output is the one we actually expect
			if diff := cmp.Diff(tc.output, output); diff != "" {
				t.Error(diff)
			}
		})
	}
}
