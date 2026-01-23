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
	startTime, endTime := lsDaysToTimeInterval(days)

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

func lsDaysToTimeInterval(days int64) (startTime, endTime time.Time) {
	now := time.Now()
	year, month, day := now.Date()
	endTime = time.Date(year, month, day, 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	daysClamped := int(min(max(0, days), 365))
	startTime = endTime.AddDate(0, 0, -daysClamped)
	return
}
