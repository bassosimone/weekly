// main.go - Main function
// SPDX-License-Identifier: GPL-3.0-or-later

// Package cli contains the CLI implementation
package cli

import (
	"context"
	"io"
	"io/fs"
	"os"
	"runtime/debug"

	"github.com/bassosimone/vclip"
	"github.com/bassosimone/vflag"
	"github.com/bassosimone/weekly/internal/calendarapi"
	"github.com/rogpeppe/go-internal/lockedfile"
)

// execEnv is the execution environment used by this tool.
type execEnv struct {
	// Args is initialized with [os.Args].
	Args []string

	// Exit is initialized with [os.Exit].
	Exit func(exitcode int)

	// LockedfileRead is initialized with [lockedfile.Read].
	LockedfileRead func(path string) ([]byte, error)

	// LockedfileWrite is initialized with [lockedfile.Write].
	LockedfileWrite func(path string, content io.Reader, perms fs.FileMode) error

	// lookupEnv is initialized with [os.LookupEnv].
	lookupEnv func(key string) (string, bool)

	// NewCalendarClient constructs a new [calendarapi.Client].
	NewCalendarClient func(ctx context.Context, credentialsPath string) (calendarapi.Client, error)

	// Stderr is initialized with [os.Stderr].
	Stderr io.Writer

	// OSStdout is initialized with [os.Stdout].
	Stdout io.Writer

	// OSStdin is initialized with [os.Stdin].
	Stdin io.Reader
}

// newExecEnv constructs a new instance of [*execEnv].
func newExecEnv() *execEnv {
	return &execEnv{
		Args:              os.Args,
		Exit:              os.Exit,
		LockedfileRead:    lockedfile.Read,
		LockedfileWrite:   lockedfile.Write,
		lookupEnv:         os.LookupEnv,
		NewCalendarClient: calendarapi.NewClient,
		Stderr:            os.Stderr,
		Stdout:            os.Stdout,
		Stdin:             os.Stdin,
	}
}

// LookupEnv calls the lookupEnv function.
func (ee *execEnv) LookupEnv(key string) (string, bool) {
	return ee.lookupEnv(key)
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
	// Create the dispatcher command
	disp := vclip.NewDispatcherCommand("weekly", vflag.ExitOnError)
	disp.AddDescription("Track weekly activity using Google Calendar.")
	disp.AddVersionHandlers(version)

	// Not strictly needed in production but necessary for testing
	disp.Exit = env.Exit
	disp.Stdout = env.Stdout
	disp.Stderr = env.Stderr

	// Create the `init` leaf command
	disp.AddCommand("init", vclip.CommandFunc(initMain), initBriefDescription)

	// Create the `ls` leaf command
	disp.AddCommand("ls", vclip.CommandFunc(lsMain), lsBriefDescription)

	// Create the `tutorial` leaf command
	disp.AddCommand("tutorial", vclip.CommandFunc(tutorialMain), tutorialBriefDescription)

	// Execute the root command
	vclip.Main(context.Background(), disp, env.Args[1:])
}
