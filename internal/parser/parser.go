// Package parser contains code to parse events.
package parser

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
)

// Event is a weekly calendar entry.
type Event struct {
	// Organization is the event $organization.
	Organization string

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

type simplifiedCalendarEvent struct {
	// Summary is the calendar event summary
	Summary string

	// StartTime is the calendar event start time
	StartTime string

	// EndTime is the calendar event end time
	EndTime string
}

func newSimplifiedCalendarEvent(ev *calendar.Event) *simplifiedCalendarEvent {
	return &simplifiedCalendarEvent{
		Summary:   ev.Summary,
		StartTime: ev.Start.DateTime,
		EndTime:   ev.End.DateTime,
	}
}

func (ev *simplifiedCalendarEvent) String() string {
	// Note: json.Marshal cannot fail for this structure
	data, _ := json.Marshal(ev)
	return string(data)
}

// Parse parses the fetched [*calendar.Event] returning [Event] entries.
func Parse(inputs []*calendar.Event) ([]Event, error) {
	outputs := make([]Event, 0, len(inputs))

	for _, input := range inputs {
		var e Event
		if err := e.parseAll(newSimplifiedCalendarEvent(input)); err != nil {
			return nil, err
		}
		outputs = append(outputs, e)
	}

	return outputs, nil
}

func (e *Event) parseAll(ev *simplifiedCalendarEvent) error {
	// Summary
	if err := e.parseSummary(ev); err != nil {
		return err
	}

	// Times
	return e.parseTimes(ev)
}

func (e *Event) parseSummary(ev *simplifiedCalendarEvent) error {
	tokens := strings.Split(ev.Summary, " ")
	if len(tokens) <= 0 {
		return fmt.Errorf("empty summary in %s", ev)
	}

	for _, token := range tokens {

		// Parse organization
		if organization, found := strings.CutPrefix(token, "$"); found {
			if e.Organization != "" {
				return fmt.Errorf("multiple organizations in %s", ev)
			}
			e.Organization = organization
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

		// Persons
		if person, found := strings.CutPrefix(token, "@"); found {
			e.Persons = append(e.Persons, person)
			continue
		}

		return fmt.Errorf("invalid token %q in %s", token, ev)
	}

	return nil
}

func parseTimeInto(output *time.Time, input string) error {
	const format = "2006-01-02T15:04:05-07:00"
	tx, err := time.Parse(format, input)
	if err != nil {
		return err
	}
	*output = tx
	return nil
}

func (e *Event) parseTimes(ev *simplifiedCalendarEvent) error {
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
