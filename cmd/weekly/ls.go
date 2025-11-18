// ls.go - ls subcommand
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bassosimone/clip"
	"github.com/bassosimone/clip/pkg/assert"
	"github.com/bassosimone/clip/pkg/nflag"
	"github.com/bassosimone/weekly/internal/calendarapi"
	"github.com/bassosimone/weekly/internal/parser"
)

// lsMain is the main entry point for the ls leaf command.
func lsMain(ctx context.Context, args *clip.CommandArgs[*clip.StdlibExecEnv]) error {
	// Create flag set
	fset := nflag.NewFlagSet(args.CommandName, nflag.ExitOnError)
	fset.Description = args.Command.BriefDescription()
	fset.PositionalArgumentsUsage = ""
	fset.MinPositionalArgs = 0
	fset.MaxPositionalArgs = 0

	// Not strictly needed in production but necessary for testing
	fset.Exit = args.Env.Exit
	fset.Stderr = args.Env.Stderr()
	fset.Stdout = args.Env.Stdout()

	// Create default values for flags
	var (
		dataDir = filepath.Join(".", "private")
		days    = int64(1)
		format  = "json"
	)

	// Add the --data-dir flag
	fset.StringFlagVar(&dataDir, "data-dir", 0, "Directory containing the configuration.")

	// Add the --days flag
	fset.Int64FlagVar(&days, "days", 0, "Number of days in the past to fetch.")

	// Add the --format flag
	fset.StringFlagVar(&format, "format", 0, "Format to emit output")

	// Add the --help flag
	fset.AutoHelp("help", 'h', "Print this help message and exit.")

	// Parse the flags
	assert.NotError(fset.Parse(args.Args))

	// Create calendar API client
	client := must1(calendarapi.NewClient(ctx, credentialsPath(dataDir)))

	// Load the calendar ID to use
	cinfo := must1(readCalendarInfo(calendarPath(dataDir)))

	// Compute start time and end time
	startTime, endTime := lsDaysToTimeInterval(days)

	// Fetch and parse the events as weekly-calendar events
	config := calendarapi.FetchEventsConfig{
		CalendarID: cinfo.ID,
		StartTime:  startTime,
		EndTime:    endTime,
	}
	rawEvents := must1(client.FetchEvents(ctx, &config))
	events := must1(parser.Parse(rawEvents))

	// TODO(bassosimone): add support for grouping events together

	// Format and print the weekly-calendar events
	lsFormat(format, os.Stdout, events)
	return nil
}

func lsFormat(format string, w io.Writer, events []parser.Event) {
	switch format {
	default:
		fallthrough
	case "json":
		lsFormatJSON(w, events)

	case "csv":
		lsFormatCSV(w, events)
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
			ev.Organization,
			ev.Activity,
			strings.Join(ev.Tags, " "),
			strings.Join(ev.Persons, " "),
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
