// main_test.go - Main function tests
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/bassosimone/weekly/internal/calendarapi"
	"github.com/google/go-cmp/cmp"
)

// outputCapturer captures the program output.
//
// The zero value is ready to use.
type outputCapturer struct {
	// buf is the output buffer
	buf bytes.Buffer

	// mu provides mutual exclusion
	mu sync.Mutex
}

var _ io.Writer = &outputCapturer{}

// Write implements [io.Writer].
func (w *outputCapturer) Write(data []byte) (int, error) {
	w.mu.Lock()
	count, err := w.buf.Write(data)
	w.mu.Unlock()
	return count, err
}

// Lines returns the captured lines.
func (w *outputCapturer) Lines() []string {
	w.mu.Lock()
	out := strings.Split(w.buf.String(), "\n")
	w.mu.Unlock()
	return out
}

// filesys abstracts the file system for the purpose of testing.
//
// The zero value is ready to use.
type filesys struct {
	// mu provides mutual exclusion
	mu sync.Mutex

	// root maps file names to their content.
	root map[string][]byte
}

// LockedfileRead atomically reads the content of a file.
func (fsx *filesys) LockedfileRead(path string) ([]byte, error) {
	fsx.mu.Lock()
	var err error
	data, found := fsx.root[path]
	if !found {
		err = fmt.Errorf("%s: %w", path, fs.ErrNotExist)
	}
	fsx.mu.Unlock()
	return data, err
}

// LockedfileWrite atomically writes the content of a file.
func (fsx *filesys) LockedfileWrite(path string, content io.Reader, perms fs.FileMode) error {
	fsx.mu.Lock()
	if fsx.root == nil {
		fsx.root = make(map[string][]byte)
	}
	data, err := io.ReadAll(content)
	if err != nil {
		fsx.mu.Unlock()
		return err
	}
	fsx.root[path] = data
	fsx.mu.Unlock()
	return nil
}

// Files returns the paths of the files inside the filesystem.
func (fsx *filesys) Files() (paths []string) {
	fsx.mu.Lock()
	paths = append(paths, slices.Sorted(maps.Keys(fsx.root))...)
	fsx.mu.Unlock()
	return
}

// calendarClient implements [calendarapi.Client].
type calendarClient struct {
	// fetchEvents returns either mocked events or an error.
	//
	// You MUST initialize this field.
	fetchEvents func(ctx context.Context, config *calendarapi.FetchEventsConfig) ([]calendarapi.Event, error)
}

var _ calendarapi.Client = &calendarClient{}

// FetchEvents implements [calendarapi.Client].
func (c *calendarClient) FetchEvents(ctx context.Context, config *calendarapi.FetchEventsConfig) ([]calendarapi.Event, error) {
	return c.fetchEvents(ctx, config)
}

var expectedWeeklyHelpOutput = []string{
	"",
	"Usage",
	"",
	"    weekly -h [args...]",
	"    weekly --help [args...]",
	"    weekly help [args...]",
	"",
	"    weekly init [args...]",
	"",
	"    weekly ls [args...]",
	"",
	"    weekly tutorial [args...]",
	"",
	"    weekly --version [args...]",
	"    weekly version [args...]",
	"",
	"Description",
	"",
	"    Track weekly activity using Google Calendar.",
	"",
	"Commands",
	"",
	"    -h, --help, help",
	"",
	"        Show help about this command or about a subcommand.",
	"",
	"    init",
	"",
	"        Initialize and select the calendar to use.",
	"",
	"    ls",
	"",
	"        List events from the selected calendar.",
	"",
	"    tutorial",
	"",
	"        Show detailed tutorial explaining the tool usage.",
	"",
	"    --version, version",
	"",
	"        Show the version number and exit.",
	"",
	"Hints",
	"",
	"    Use `weekly <command> --help' to get command-specific help.",
	"",
	"    Append `--help' or `-h' to any command line failing with usage",
	"    errors to hide the error and obtain contextual help.",
	"",
	"",
}

var expectedWeeklyHelpLsOutput = []string{
	"",
	"Usage",
	"",
	"    weekly ls [flags]",
	"",
	"Description",
	"",
	"    List events from the selected calendar.",
	"",
	"Flags",
	"",
	"    --aggregate STRING (default: ``)",
	"",
	"        Aggregate entries (daily, weekly, or monthly).",
	"",
	"    --config-dir STRING (default: `weekly`)",
	"",
	"        Directory containing the configuration.",
	"",
	"    --days INT64 (default: `1`)",
	"",
	"        Number of days in the past to fetch.",
	"",
	"    --format STRING (default: `box`)",
	"",
	"        Format to emit output: box, csv, invoice, json.",
	"",
	"    -h, --help",
	"",
	"        Print this help message and exit.",
	"",
	"    --max-events INT64 (default: `4096`)",
	"",
	"        Set maximum number of events to fetch.",
	"",
	"    --project STRING (default: ``)",
	"",
	"        Only show data for the given project.",
	"",
	"    --tag STRING (default: ``)",
	"",
	"        Only show data for the given tag.",
	"",
	"    --total[=BOOL] (default: `false`)",
	"",
	"        Compute total amount of hours worked.",
	"",
	"Examples",
	"",
	"    To see what you have done today in a user friendly format use:",
	"",
	"        weekly ls",
	"",
	"    To get the same data in a format suitable for invoicing:",
	"",
	"        weekly ls --format invoice --aggregate daily",
	"",
	"    You can also change the format to be JSON:",
	"",
	"        weekly ls --format json",
	"",
	"    Alternatively, you can change the format to be CSV:",
	"",
	"        weekly ls --format csv",
	"",
	"    You can go back in time with the `--days` flag:",
	"",
	"        weekly ls --days 3",
	"",
	"    You can aggregate daily and by project with `--aggregate`:",
	"",
	"        weekly ls --days 3 --aggregate daily",
	"",
	"    You can also aggregate weekly:",
	"",
	"        weekly ls --days 7 --aggregate weekly",
	"",
	"    You can also aggregate monthly:",
	"",
	"        weekly ls --days 60 --aggregate monthly",
	"",
	"    You can also compute the total in the aggregation period:",
	"",
	"        weekly ls --total",
	"",
	"    You can also filter by project:",
	"",
	"        weekly ls --project mlab",
	"",
	"    You can also filter by tag:",
	"",
	"        weekly ls --tag neubot",
	"",
	"    The `invoice` format is a simplified CSV format suitable for",
	"    generating invoices.",
	"",
	"",
}

var expectedWeeklyHelpInitOutput = []string{
	"",
	"Usage",
	"",
	"    weekly init [flags]",
	"",
	"Description",
	"",
	"    Initialize and select the calendar to use.",
	"",
	"Flags",
	"",
	"    --config-dir STRING (default: `weekly`)",
	"",
	"        Directory containing the configuration.",
	"",
	"    -h, --help",
	"",
	"        Print this help message and exit.",
	"",
	"",
}

var expectedWeeklyHelpTutorialOutput = []string{
	"",
	"Usage",
	"",
	"    weekly tutorial [flags]",
	"",
	"Description",
	"",
	"    Show detailed tutorial explaining the tool usage.",
	"",
	"Flags",
	"",
	"    -h, --help",
	"",
	"        Print this help message and exit.",
	"",
	"",
}

func TestMain(t *testing.T) {
	// describes a test case run by this function
	type testCase struct {
		// filesBefore contains the file system state
		// before executing the Main function.
		filesBefore map[string][]byte

		// eventsToReturn optionally contains the events to
		// return from calendarapi.Client.FetchEvents.
		//
		// If this field is not nil, we use mocks to simulate
		// receiving this events from the API call.
		eventsToReturn []calendarapi.Event

		// argv contains the command line
		argv []string

		// stdin contains the data available on the stdin
		stdin io.Reader

		// stdoutLines contains the expected lines on the stdout
		stdoutLines []string

		// stderrLines contains the expected lines on the stderr
		stderrLines []string

		// exitCode contains the expected exit code.
		exitCode int64

		// modifiedFiles contains the files that we expect
		// to see to be modified by the test iself.
		modifiedFiles map[string][]byte
	}

	// defines all test cases
	cases := []testCase{

		// ====================================================-
		// Root Command
		// ====================================================-

		// `weekly` should print the help screen
		{
			argv:        []string{"weekly"},
			stdoutLines: expectedWeeklyHelpOutput,
			stderrLines: []string{""},
		},

		// `weekly -h` should print the help screen
		{
			argv:        []string{"weekly", "-h"},
			stdoutLines: expectedWeeklyHelpOutput,
			stderrLines: []string{""},
		},

		// `weekly --help` should print the help screen
		{
			argv:        []string{"weekly", "--help"},
			stdoutLines: expectedWeeklyHelpOutput,
			stderrLines: []string{""},
		},

		// `weekly help` should print the help screen
		{
			argv:        []string{"weekly", "help"},
			stdoutLines: expectedWeeklyHelpOutput,
			stderrLines: []string{""},
		},

		// `weekly --version` should print the program version
		{
			argv: []string{"weekly", "--version"},
			stdoutLines: []string{
				version,
				"",
			},
			stderrLines: []string{""},
		},

		// `weekly version` command should print the program version
		{
			argv: []string{"weekly", "version"},
			stdoutLines: []string{
				version,
				"",
			},
			stderrLines: []string{""},
		},

		// `weekly --invalid-flag` should print an error
		{
			argv:        []string{"weekly", "--invalid-flag"},
			stdoutLines: []string{""},
			stderrLines: []string{
				"weekly: command not found: --invalid-flag",
				"hint: use `weekly --help' to see the available commands",
				"",
			},
			exitCode: 2,
		},

		// `weekly invalid-command` should print an error
		{
			argv:        []string{"weekly", "invalid-command"},
			stdoutLines: []string{""},
			stderrLines: []string{
				"weekly: command not found: invalid-command",
				"hint: use `weekly --help' to see the available commands",
				"",
			},
			exitCode: 2,
		},

		// ====================================================-
		// `tutorial` Command
		// ====================================================-

		// `weekly tutorial --help` should print the help screen
		{
			argv:        []string{"weekly", "tutorial", "--help"},
			stdoutLines: expectedWeeklyHelpTutorialOutput,
			stderrLines: []string{""},
			exitCode:    0,
		},

		// `weekly tutorial -h` should print the help screen
		{
			argv:        []string{"weekly", "tutorial", "-h"},
			stdoutLines: expectedWeeklyHelpTutorialOutput,
			stderrLines: []string{""},
			exitCode:    0,
		},

		// `weekly help tutorial` should print the help screen
		{
			argv:        []string{"weekly", "tutorial", "-h"},
			stdoutLines: expectedWeeklyHelpTutorialOutput,
			stderrLines: []string{""},
			exitCode:    0,
		},

		// `weekly tutorial` should print the tutorial screen
		{
			argv:        []string{"weekly", "tutorial"},
			stdoutLines: strings.Split(tutorialData, "\n"),
			stderrLines: []string{""},
			exitCode:    0,
		},

		// ====================================================-
		// `init` Command
		// ====================================================-

		// `weekly init --help` should print the help screen
		{
			argv:        []string{"weekly", "init", "--help"},
			stdoutLines: expectedWeeklyHelpInitOutput,
			stderrLines: []string{""},
			exitCode:    0,
		},

		// `weekly init -h` should print the help screen
		{
			argv:        []string{"weekly", "init", "-h"},
			stdoutLines: expectedWeeklyHelpInitOutput,
			stderrLines: []string{""},
			exitCode:    0,
		},

		// `weekly help init` should print the help screen
		{
			argv:        []string{"weekly", "help", "init"},
			stdoutLines: expectedWeeklyHelpInitOutput,
			stderrLines: []string{""},
			exitCode:    0,
		},

		// `weekly init` should initialize the XDG_DATA_HOME dir
		{
			argv:  []string{"weekly", "init"},
			stdin: strings.NewReader("0xdeadbeef"),
			stdoutLines: []string{
				"Please, provide the default calendar ID: ",
			},
			stderrLines: []string{""},
			exitCode:    0,
			modifiedFiles: map[string][]byte{
				"weekly/calendar.json": []byte(`{"ID":"0xdeadbeef"}`),
			},
		},

		// ====================================================-
		// `ls` Command
		// ====================================================-

		// `weekly ls --help` should print the help screen
		{
			argv:        []string{"weekly", "ls", "--help"},
			stdoutLines: expectedWeeklyHelpLsOutput,
			stderrLines: []string{""},
			exitCode:    0,
		},

		// `weekly ls -h` should print the help screen
		{
			argv:        []string{"weekly", "ls", "-h"},
			stdoutLines: expectedWeeklyHelpLsOutput,
			stderrLines: []string{""},
			exitCode:    0,
		},

		// `weekly help ls` should print the help screen
		{
			argv:        []string{"weekly", "help", "ls"},
			stdoutLines: expectedWeeklyHelpLsOutput,
			stderrLines: []string{""},
			exitCode:    0,
		},

		// `weekly ls --format invoice` should return some events
		{
			argv: []string{"weekly", "ls", "--format", "invoice"},
			filesBefore: map[string][]byte{
				"weekly/calendar.json": []byte(`{"ID":"0xdeadbeef"}`),
			},
			eventsToReturn: []calendarapi.Event{
				{
					Summary:   "$nexa %development #neubot",
					StartTime: "2016-12-08T10:00:00+01:00",
					EndTime:   "2016-12-08T13:00:00+01:00",
				},
				{
					Summary:   "$nexa %development #neubot",
					StartTime: "2016-12-08T15:30:00+01:00",
					EndTime:   "2016-12-08T17:00:00+01:00",
				},
				{
					Summary:   "$nexa %meeting #wednesday",
					StartTime: "2016-12-08T18:00:00+01:00",
					EndTime:   "2016-12-08T20:00:00+01:00",
				},
			},
			stdoutLines: []string{
				"nexa,2016-12-08,3",
				"nexa,2016-12-08,1.5",
				"nexa,2016-12-08,2",
				"",
			},
			stderrLines: []string{""},
			exitCode:    0,
		},

		// `weekly ls --format invoice --max-events 3` should also warn
		{
			argv: []string{"weekly", "ls", "--format", "invoice", "--max-events", "3"},
			filesBefore: map[string][]byte{
				"weekly/calendar.json": []byte(`{"ID":"0xdeadbeef"}`),
			},
			eventsToReturn: []calendarapi.Event{
				{
					Summary:   "$nexa %development #neubot",
					StartTime: "2016-12-08T10:00:00+01:00",
					EndTime:   "2016-12-08T13:00:00+01:00",
				},
				{
					Summary:   "$nexa %development #neubot",
					StartTime: "2016-12-08T15:30:00+01:00",
					EndTime:   "2016-12-08T17:00:00+01:00",
				},
				{
					Summary:   "$nexa %meeting #wednesday",
					StartTime: "2016-12-08T18:00:00+01:00",
					EndTime:   "2016-12-08T20:00:00+01:00",
				},
			},
			stdoutLines: []string{
				"nexa,2016-12-08,3",
				"nexa,2016-12-08,1.5",
				"nexa,2016-12-08,2",
				"",
			},
			stderrLines: []string{
				"warning: reached maximum number of events to query (3)",
				"warning: try increasing the limit using `--max-events`",
				"",
			},
			exitCode: 0,
		},
	}

	// run each test case
	for _, tc := range cases {
		t.Run(strings.Join(tc.argv, " "), func(t *testing.T) {
			// Save and restore the global env
			oldEnv := env
			defer func() {
				env = oldEnv
			}()

			// replace and edit the test environment
			env = newExecEnv()
			env.Args = tc.argv

			env.Stdin = tc.stdin

			stdout := &outputCapturer{}
			env.Stdout = stdout

			stderr := &outputCapturer{}
			env.Stderr = stderr

			errPanicSentinel := errors.New("panic invoked")
			exitCode := &atomic.Int64{}
			env.Exit = func(code int) {
				exitCode.Store(int64(code))
				panic(errPanicSentinel)
			}

			beforeFS := &filesys{
				mu:   sync.Mutex{},
				root: tc.filesBefore, // make before files available
			}
			env.LockedfileRead = beforeFS.LockedfileRead

			afterFS := &filesys{} // zero value is OK
			env.LockedfileWrite = afterFS.LockedfileWrite

			env.lookupEnv = func(key string) (string, bool) {
				if key == "XDG_CONFIG_HOME" {
					return ".", true
				}
				return "", false
			}

			if len(tc.eventsToReturn) >= 1 {
				env.NewCalendarClient = func(ctx context.Context, path string) (calendarapi.Client, error) {
					c := &calendarClient{
						fetchEvents: func(ctx context.Context,
							config *calendarapi.FetchEventsConfig) ([]calendarapi.Event, error) {
							return tc.eventsToReturn, nil
						},
					}
					return c, nil
				}
			}

			// execute the function to test
			func() {
				// carefully handle panics inside Main
				defer func() {
					if r := recover(); r != nil {
						err, ok := r.(error)
						if !ok {
							t.Error("unexpected panic", r)
							return
						}
						if !errors.Is(err, errPanicSentinel) {
							t.Error("unexpected panic", r)
							return
						}
						// all good: this panic was caused by the
						// mocked [os.Exit] we did setup above
					}
				}()

				// invoke the function we are testing
				Main()
			}()

			// make sure the stdout is as expected
			if diff := cmp.Diff(tc.stdoutLines, stdout.Lines()); diff != "" {
				t.Error("stdout differs:", diff)
			}

			// make sure the stderr is as expected
			if diff := cmp.Diff(tc.stderrLines, stderr.Lines()); diff != "" {
				t.Error("stderr differs:", diff)
			}

			// make sure exitcode is as expected
			if diff := cmp.Diff(tc.exitCode, exitCode.Load()); diff != "" {
				t.Error("exit code differs:", diff)
			}

			// check all the files that have been modified
			expectModifiedFilesNames := slices.Sorted(maps.Keys(tc.modifiedFiles))
			gotModifiedFilesNames := afterFS.Files()
			if diff := cmp.Diff(expectModifiedFilesNames, gotModifiedFilesNames); diff != "" {
				t.Error("expected files differ:", diff)
			}

			for _, path := range gotModifiedFilesNames {
				expectData, ok := tc.modifiedFiles[path]
				if !ok {
					t.Error("file named", path, "written but not expected")
					continue
				}

				gotData, err := afterFS.LockedfileRead(path)
				if err != nil {
					t.Error("file named", path, "expected but not written", err)
					continue
				}

				if diff := cmp.Diff(expectData, gotData); diff != "" {
					t.Error("expected data differ for the", path, "file:", diff)
					continue
				}
			}
		})
	}
}
