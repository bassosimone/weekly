// ls.go - ls subcommand
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/bassosimone/clip"
	"github.com/bassosimone/clip/pkg/assert"
	"github.com/bassosimone/clip/pkg/nflag"
	"github.com/bassosimone/weekly/internal/calendarapi"
	"github.com/bassosimone/weekly/internal/parser"
	"github.com/bassosimone/weekly/internal/pipeline"
	"github.com/olekukonko/tablewriter"
)

//go:embed docs/ls_examples.txt
var lsExamplesTxt string

// lsMain is the main entry point for the `ls` leaf command.
func lsMain(ctx context.Context, args *clip.CommandArgs[*clip.StdlibExecEnv]) error {
	// Create flag set
	fset := nflag.NewFlagSet(args.CommandName, nflag.ExitOnError)
	fset.Description = args.Command.BriefDescription()
	fset.PositionalArgumentsUsage = ""
	fset.MinPositionalArgs = 0
	fset.MaxPositionalArgs = 0
	fset.Examples = lsExamplesTxt

	// Not strictly needed in production but necessary for testing
	fset.Exit = args.Env.Exit
	fset.Stderr = args.Env.Stderr()
	fset.Stdout = args.Env.Stdout()

	// Create default values for flags
	var (
		configDir = xdgConfigHome(args.Env)
		days      = int64(1)
		format    = "box"
		maxEvents = int64(4096)
		pconfig   = pipeline.Config{
			Aggregate: "",
			Project:   "",
			Total:     false,
		}
	)

	// Add the --aggregate
	fset.StringFlagVar(&pconfig.Aggregate, "aggregate", 0, "Aggregate entries (daily or monthly).")

	// Add the --config-dir flag
	fset.StringFlagVar(&configDir, "config-dir", 0, "Directory containing the configuration.")

	// Add the --days flag
	fset.Int64FlagVar(&days, "days", 0, "Number of days in the past to fetch.")

	// Add the --format flag
	fset.StringFlagVar(&format, "format", 0, "Format to emit output: box (default), csv, invoice, json.")

	// Add the --help flag
	fset.AutoHelp("help", 'h', "Print this help message and exit.")

	// Add the --max-events flag
	fset.Int64FlagVar(&maxEvents, "max-events", 0, "Set maximum number of events to fetch.")

	// Add the --project flag
	fset.StringFlagVar(&pconfig.Project, "project", 0, "Only show data for the given project.")

	// Add the --total flag
	fset.BoolFlagVar(&pconfig.Total, "total", 0, "Compute total amount of hours worked.")

	// Parse the flags
	assert.NotError(fset.Parse(args.Args))

	// Create calendar API client
	client := must1(calendarapi.NewClient(ctx, credentialsPath(configDir)))

	// Load the calendar ID to use
	cinfo := must1(readCalendarInfo(calendarPath(configDir)))

	// Compute start time and end time
	startTime, endTime := lsDaysToTimeInterval(days)

	// Fetch and parse the events as weekly-calendar events
	config := calendarapi.FetchEventsConfig{
		CalendarID: cinfo.ID,
		StartTime:  startTime,
		EndTime:    endTime,
		MaxEvents:  maxEvents,
	}
	rawEvents := must1(client.FetchEvents(ctx, &config))
	events := must1(parser.Parse(rawEvents))

	// Maybe emit warning depending on the number of events
	lsMaybeWarnOnEventsNumber(maxEvents, events)

	// Run the events processing pipeline
	events = must1(pipeline.Run(&pconfig, events))

	// Format and print the weekly-calendar events
	lsFormat(format, os.Stdout, events)
	return nil
}

func lsMaybeWarnOnEventsNumber(maxEvents int64, events []parser.Event) {
	if int64(len(events)) >= maxEvents {
		fmt.Fprintf(os.Stderr, "warning: reached maximum number of events to query (%d)\n", maxEvents)
		fmt.Fprintf(os.Stderr, "warning: try increasing the limit using `--max-events`\n")
	}
}

func lsFormat(format string, w io.Writer, events []parser.Event) {
	switch format {
	case "box":
		lsFormatBox(w, events)

	case "csv":
		lsFormatCSV(w, events)

	case "invoice":
		lsFormatInvoice(w, events)

	case "json":
		lsFormatJSON(w, events)

	default:
		must0(errors.New("the --format flag accepts one of these values: box, csv, invoice, json"))
	}
}

func lsFormatJSON(w io.Writer, events []parser.Event) {
	for _, ev := range events {
		_ = must1(fmt.Fprintf(w, "%s\n", string(must1(json.Marshal(ev)))))
	}
}

func lsFormatCSV(w io.Writer, events []parser.Event) {
	cw := csv.NewWriter(w)
	for _, ev := range events {
		cw.Write([]string{
			ev.StartTime.Format(time.RFC3339),
			ev.Duration.String(),
			ev.Project,
			ev.Activity,
			strings.Join(ev.Tags, " "),
			strings.Join(ev.Persons, " "),
		})
	}
	cw.Flush()
	must0(cw.Error())
}

func lsFormatBox(w io.Writer, events []parser.Event) {
	data := [][]any{
		{"StartTime", "Hours", "Project", "Activity", "Tags", "Persons"},
	}
	for _, ev := range events {
		data = append(data, []any{
			ev.StartTime.Format("2006-01-02 15:04"),
			fmt.Sprintf("%6.1f", ev.Duration.Hours()),
			ev.Project,
			ev.Activity,
			strings.Join(ev.Tags, " "),
			strings.Join(ev.Persons, " "),
		})
	}

	table := tablewriter.NewTable(w)
	table.Header(data[0])
	table.Bulk(data[1:])
	must0(table.Render())
}

func lsFormatInvoice(w io.Writer, events []parser.Event) {
	cw := csv.NewWriter(w)
	for _, ev := range events {
		cw.Write([]string{
			ev.Project,
			ev.StartTime.Format("2006-01-02"),
			fmt.Sprint(ev.Duration.Hours()),
		})
	}
	cw.Flush()
	must0(cw.Error())
}

func lsDaysToTimeInterval(days int64) (startTime, endTime time.Time) {
	now := time.Now()
	year, month, day := now.Date()
	endTime = time.Date(year, month, day, 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	daysClamped := int(min(max(0, days), 365))
	startTime = endTime.AddDate(0, 0, -daysClamped)
	return
}
