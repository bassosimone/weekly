// main.go - Main function
// SPDX-License-Identifier: GPL-3.0-or-later

// Package cli contains the CLI implementation
package cli

import (
	"context"
	"io"
	"io/fs"
	"runtime/debug"

	"github.com/bassosimone/clip"
	"github.com/bassosimone/clip/pkg/nflag"
	"github.com/bassosimone/weekly/internal/calendarapi"
	"github.com/rogpeppe/go-internal/lockedfile"
)

// execEnv is the execution environment used by this tool.
type execEnv struct {
	// We embed a [*clip.StdlibExecEnv]
	*clip.StdlibExecEnv

	// lockedfileRead allows mocking calls to [lockedfile.Read].
	lockedfileRead func(path string) ([]byte, error)

	// lockedfileWrite allows mocking calls to [lockedfile.Write].
	lockedfileWrite func(path string, content io.Reader, perms fs.FileMode) error

	// newCalendarClient constructs a new [calendarapi.Client].
	newCalendarClient func(ctx context.Context, credentialsPath string) (calendarapi.Client, error)
}

var _ clip.ExecEnv = &execEnv{}

// newExecEnv constructs a new instance of [*execEnv].
func newExecEnv() *execEnv {
	return &execEnv{
		StdlibExecEnv:     clip.NewStdlibExecEnv(),
		lockedfileRead:    lockedfile.Read,
		lockedfileWrite:   lockedfile.Write,
		newCalendarClient: calendarapi.NewClient,
	}
}

// LockedfileRead is equivalent to [lockedfile.Read].
func (env *execEnv) LockedfileRead(path string) ([]byte, error) {
	return env.lockedfileRead(path)
}

// LockedfileWrite is equivalent to [lockedfile.Write].
func (env *execEnv) LockedfileWrite(path string, content io.Reader, perms fs.FileMode) error {
	return env.lockedfileWrite(path, content, perms)
}

// NewCalendarClient constructs a new [calendarapi.Client] instance.
func (env *execEnv) NewCalendarClient(ctx context.Context, credentialsPath string) (calendarapi.Client, error) {
	return env.newCalendarClient(ctx, credentialsPath)
}

var (
	// env is the global execution environment used throughout the CLI.
	//
	// This is intentionally global and mutable to enable comprehensive testing.
	// Tests replace env entirely to mock all dependencies (filesystem, network,
	// exit behavior, etc.) without requiring complex dependency injection.
	//
	// See main_test.go for the testing pattern: each test saves the original env,
	// creates a fresh test environment with mocked dependencies, runs the code,
	// and restores the original env via defer.
	//
	// While global mutable state is generally avoided, this is appropriate for
	// a CLI application where:
	//   - There is a single main execution path (not a library used by others)
	//   - Testing requires complete control over all side effects
	//   - The alternative would be threading env through every function call
	env = newExecEnv()

	// version contains the program version string.
	//
	// This is set during init() from build information.
	version string
)

func init() {
	// Define the overall suite version
	version = "(devel)"
	if binfo, ok := debug.ReadBuildInfo(); ok {
		version = binfo.Main.Version
	}
}

// Main is the main function of the CLI implementation.
func Main() {
	// Create the `init` leaf command
	initCmd := &clip.LeafCommand[*execEnv]{
		BriefDescriptionText: "Initialize and select the calendar to use.",
		RunFunc:              initMain,
	}

	// Create the `ls` leaf command
	lsCmd := &clip.LeafCommand[*execEnv]{
		BriefDescriptionText: "List events from the selected calendar.",
		RunFunc:              lsMain,
	}

	// Create the `tutorial` leaf command
	tutorialCmd := &clip.LeafCommand[*execEnv]{
		BriefDescriptionText: "Show detailed tutorial explaining the tool usage.",
		RunFunc:              tutorialMain,
	}

	// Create the root command
	rootCmd := &clip.RootCommand[*execEnv]{
		Command: &clip.DispatcherCommand[*execEnv]{
			BriefDescriptionText: "Track weekly activity using Google Calendar.",
			Commands: map[string]clip.Command[*execEnv]{
				"init":     initCmd,
				"ls":       lsCmd,
				"tutorial": tutorialCmd,
			},
			ErrorHandling:             nflag.ExitOnError,
			Version:                   version,
			OptionPrefixes:            []string{"-", "--"},
			OptionsArgumentsSeparator: "--",
		},
		AutoCancel: true,
	}

	// Execute the root command
	rootCmd.Main(env)
}
