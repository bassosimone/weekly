// output.go - Code that emits output
// SPDX-License-Identifier: GPL-3.0-or-later

// Package output contains code to output weekly events.
package output

import (
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/bassosimone/weekly/internal/parser"
	"github.com/olekukonko/tablewriter"
)

// Write writes the events using the given writer and output format.
func Write(w io.Writer, format string, events []parser.Event) error {
	switch format {
	case "box":
		return writeFormatBox(w, events)

	case "csv":
		return writeFormatCSV(w, events)

	case "invoice":
		return writeFormatInvoice(w, events)

	case "json":
		return writeFormatJSON(w, events)

	default:
		return errors.New("the --format flag accepts one of these values: box, csv, invoice, json")
	}
}

func writeFormatJSON(w io.Writer, events []parser.Event) error {
	for _, ev := range events {
		// Note that JSON serialization of an event cannot failt
		serialized, _ := json.Marshal(ev)
		if _, err := fmt.Fprintf(w, "%s\n", string(serialized)); err != nil {
			return err
		}
	}
	return nil
}

func writeFormatCSV(w io.Writer, events []parser.Event) error {
	cw := csv.NewWriter(w)
	for _, ev := range events {
		_ = cw.Write([]string{
			ev.StartTime.Format(time.RFC3339),
			ev.Duration.String(),
			ev.Project,
			ev.Activity,
			strings.Join(ev.Tags, " "),
			strings.Join(ev.Persons, " "),
		})
	}
	cw.Flush()
	return cw.Error()
}

func writeFormatBox(w io.Writer, events []parser.Event) error {
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
	_ = table.Bulk(data[1:]) // We do not expect a failure here
	return table.Render()
}

func writeFormatInvoice(w io.Writer, events []parser.Event) error {
	cw := csv.NewWriter(w)
	for _, ev := range events {
		_ = cw.Write([]string{
			ev.Project,
			ev.StartTime.Format("2006-01-02"),
			fmt.Sprint(ev.Duration.Hours()),
		})
	}
	cw.Flush()
	return cw.Error()
}
