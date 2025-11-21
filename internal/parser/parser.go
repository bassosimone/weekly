// parser.go - weekly-events parser implementation
// SPDX-License-Identifier: GPL-3.0-or-later

// Package parser contains code to parse events.
package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/bassosimone/weekly/internal/calendarapi"
)

// Event is a weekly calendar entry.
type Event struct {
	// Project is the event funded $project.
	Project string

	// Activity is the event %activity.
	Activity string

	// Tags contains the events #tag list.
	Tags []string

	// Persons contains the events @person list.
	Persons []string

	// StartTime is the event start time.
	StartTime time.Time

	// Duration is the event duration.
	Duration time.Duration
}

// Parse parses the fetched [*calendar.Event] returning [Event] entries.
func Parse(inputs []calendarapi.Event) ([]Event, error) {
	outputs := make([]Event, 0, len(inputs))

	for _, input := range inputs {
		e := Event{
			Project:   "",
			Activity:  "",
			Tags:      []string{},
			Persons:   []string{},
			StartTime: time.Time{},
			Duration:  0,
		}
		if err := e.parseAll(&input); err != nil {
			return nil, err
		}
		outputs = append(outputs, e)
	}

	return outputs, nil
}

func (e *Event) parseAll(ev *calendarapi.Event) error {
	// Parse summary
	if err := e.parseSummary(ev); err != nil {
		return err
	}

	// Parse times
	return e.parseTimes(ev)
}

func (e *Event) parseSummary(ev *calendarapi.Event) error {
	// Example entry: `$mlab %development #iqb @sbasso`

	for token := range strings.SplitSeq(ev.Summary, " ") {

		// Parse project
		if project, found := strings.CutPrefix(token, "$"); found {
			if e.Project != "" {
				return fmt.Errorf("multiple projects in %s", ev)
			}
			e.Project = project
			continue
		}

		// Parse activity
		if activity, found := strings.CutPrefix(token, "%"); found {
			if e.Activity != "" {
				return fmt.Errorf("multiple activities in %s", ev)
			}
			e.Activity = activity
			continue
		}

		// Parse tags
		if tag, found := strings.CutPrefix(token, "#"); found {
			e.Tags = append(e.Tags, tag)
			continue
		}

		// Parse persons
		if person, found := strings.CutPrefix(token, "@"); found {
			e.Persons = append(e.Persons, person)
			continue
		}

		// Otherwise: ignore the token
	}

	// Ensure we have a project and an activity
	if e.Project == "" || e.Activity == "" {
		return fmt.Errorf("no project or activity in %s", ev)
	}

	return nil
}

// timeFormat is the format expected for calendar time entries.
const timeFormat = "2006-01-02T15:04:05-07:00"

func parseTimeInto(output *time.Time, input string) error {
	tx, err := time.Parse(timeFormat, input)
	if err != nil {
		return err
	}

	*output = tx
	return nil
}

func (e *Event) parseTimes(ev *calendarapi.Event) error {
	if err := parseTimeInto(&e.StartTime, ev.StartTime); err != nil {
		return fmt.Errorf("invalid start time in %s: %w", ev, err)
	}

	var endTime time.Time
	if err := parseTimeInto(&endTime, ev.EndTime); err != nil {
		return fmt.Errorf("invalid end time in %s: %w", ev, err)
	}
	e.Duration = endTime.Sub(e.StartTime)

	return nil
}
