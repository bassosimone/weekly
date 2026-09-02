// ls.go - ls subcommand
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/bassosimone/runtimex"
	"github.com/bassosimone/vflag"
	"github.com/bassosimone/weekly/internal/calendarapi"
	"github.com/bassosimone/weekly/internal/output"
	"github.com/bassosimone/weekly/internal/parser"
	"github.com/bassosimone/weekly/internal/pipeline"
)

//go:embed lsexamples.txt
var lsExamplesData string

// lsBriefDescription is the `ls` leaf command brief description.
const lsBriefDescription = "List events from the selected calendar."

// lsMain is the main entry point for the `ls` leaf command.
func lsMain(ctx context.Context, args []string) error {
	// Create flag set
	fset := vflag.NewFlagSet("weekly ls", vflag.ExitOnError)
	usage := vflag.NewDefaultUsagePrinter()
	usage.AddDescription(lsBriefDescription)
	usage.AddExamples(strings.Split(lsExamplesData, "\n\n")...)
	fset.UsagePrinter = usage

	// Not strictly needed in production but necessary for testing
	fset.Exit = env.Exit
	fset.Stderr = env.Stderr
	fset.Stdout = env.Stdout

	// Create default values for flags
	var (
		asOf      = ""
		configDir = xdgConfigHome(env)
		days      = int64(1)
		format    = "box"
		maxEvents = int64(4096)
		pconfig   = pipeline.Config{
			Aggregate: "",
			Project:   "",
			Tag:       "",
			Total:     false,
		}
	)

	// Add the --aggregate flag
	fset.StringVar(
		&pconfig.Aggregate,
		0,
		"aggregate",
		"Optionally aggregate entries using a `POLICY`.",
		"If empty, there's no aggregation.",
		"Valid policies: daily, weekly, and monthly.",
		"Default: empty.",
	)

	// Add the --as-of flag
	fset.StringVar(
		&asOf,
		0,
		"as-of",
		"Use `DATE` as the most recent day to fetch, instead of today.",
		"The format is YYYY-MM-DD, in the local time zone.",
		"The whole DATE day is included in the fetched interval.",
		"Default: the current day.",
	)

	// Add the --config-dir flag
	fset.StringVar(
		&configDir,
		0,
		"config-dir",
		"Select `DIR` containing the configuration.",
		"Default: `@DEFAULT_VALUE@`.",
	)

	// Add the --days flag
	fset.Int64Var(
		&days,
		0,
		"days",
		"Number of days in the past to fetch.",
		"Default: `@DEFAULT_VALUE@`.",
	)

	// Add the --format flag
	fset.StringVar(
		&format,
		0,
		"format",
		"The `FORMAT` for formatting output.",
		"Valid values: box, csv, invoice, json.",
		"Default: `@DEFAULT_VALUE@`.",
	)

	// Add the --help flag
	fset.AutoHelp('h', "help", "Print this help message and exit.")

	// Add the --max-events flag
	fset.Int64Var(
		&maxEvents,
		0,
		"max-events",
		"Set the maximum number `N` of events to fetch.",
		"Default: `@DEFAULT_VALUE@`.",
	)

	// Add the --project flag
	fset.StringVar(
		&pconfig.Project,
		0,
		"project",
		"Only show data for the given `PROJECT`.",
	)

	// Add the --tag flag
	fset.StringVar(
		&pconfig.Tag,
		0,
		"tag",
		"Only show data for the given `TAG`.",
	)

	// Add the --total flag
	fset.BoolVar(
		&pconfig.Total,
		0,
		"total",
		"Compute total amount of hours worked.",
	)

	// Parse the flags
	runtimex.PanicOnError0(fset.Parse(args))

	// Create calendar API client
	client := runtimex.LogFatalOnError1(env.NewCalendarClient(ctx, credentialsPath(configDir)))

	// Load the calendar ID to use
	cinfo := runtimex.LogFatalOnError1(readCalendarInfo(env, calendarPath(configDir)))

	// Compute start time and end time
	anchor := runtimex.LogFatalOnError1(lsParseAsOf(asOf, time.Now()))
	startTime, endTime := lsDaysToTimeInterval(anchor, days)

	// Fetch and parse the events as weekly-calendar events
	config := calendarapi.FetchEventsConfig{
		CalendarID: cinfo.ID,
		StartTime:  startTime,
		EndTime:    endTime,
		MaxEvents:  maxEvents,
	}
	rawEvents := runtimex.LogFatalOnError1(client.FetchEvents(ctx, &config))
	events := runtimex.LogFatalOnError1(parser.Parse(rawEvents))

	// Maybe emit warning depending on the number of events
	lsMaybeWarnOnEventsNumber(maxEvents, events)

	// Run the events processing pipeline
	events = runtimex.LogFatalOnError1(pipeline.Run(&pconfig, events))

	// Format and print the weekly-calendar events
	runtimex.LogFatalOnError0(output.Write(env.Stdout, format, events))
	return nil
}

func lsMaybeWarnOnEventsNumber(maxEvents int64, events []parser.Event) {
	if int64(len(events)) >= maxEvents {
		fmt.Fprintf(env.Stderr, "warning: reached maximum number of events to query (%d)\n", maxEvents)
		fmt.Fprintf(env.Stderr, "warning: try increasing the limit using `--max-events`\n")
	}
}

// lsAsOfFormat is the format accepted by the `--as-of` flag.
const lsAsOfFormat = "2006-01-02"

// lsParseAsOf converts the `--as-of` flag value to the anchor day, i.e.,
// the most recent day belonging to the interval we are going to fetch.
//
// The value argument is the raw flag value. The empty string means that
// the user did not use the flag, in which case we return now.
//
// The now argument is the current time, whose location we also use
// to extract and apply a time zone to user-supplied values.
//
// The return value is either the anchor day or an error.
func lsParseAsOf(value string, now time.Time) (time.Time, error) {
	if value == "" {
		return now, nil
	}
	anchor, err := time.ParseInLocation(lsAsOfFormat, value, now.Location())
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid --as-of value: %w", err)
	}
	return anchor, nil
}

// lsDaysToTimeInterval maps the anchor day and the number of days to fetch
// to the interval to query, with the start included and the end excluded.
//
// The interval covers the whole anchor day plus the days-1 days preceding
// it. Hence, days == 1 selects the anchor day alone. The degenerate case
// days == 0 yields the empty time interval.
//
// We clamp days to the [0, 365] range, so that a negative or otherwise
// unreasonable value cannot turn into an unreasonable query.
func lsDaysToTimeInterval(anchor time.Time, days int64) (startTime, endTime time.Time) {
	year, month, day := anchor.Date()
	endTime = time.Date(year, month, day, 0, 0, 0, 0, anchor.Location()).AddDate(0, 0, 1)
	daysClamped := int(min(max(0, days), 365))
	startTime = endTime.AddDate(0, 0, -daysClamped)
	return
}
