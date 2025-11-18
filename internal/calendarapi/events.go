// events.go - code to fetch events
// SPDX-License-Identifier: GPL-3.0-or-later

package calendarapi

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
)

// FetchEventsConfig contains config for [*Client.FetchEvents].
//
// Initialize all MANDATORY fields.
type FetchEventsConfig struct {
	// CalendarID is the MANDATORY calendar ID to use.
	CalendarID string

	// StartTime is the MANDATORY moment in time where to start.
	StartTime time.Time

	// EndTime is the MANDATORY moment in time where to end.
	EndTime time.Time
}

// FetchEvents retrieves calendar events within the specified time range.
//
// The ctx argument allows to cancel a pending call.
//
// The calendarID argument is the string identifier of the calendar.
//
// The timeMin, timeMax arguments identify the time range.
//
// The return value is either a non-empty list or an error.
func (c *Client) FetchEvents(ctx context.Context, config *FetchEventsConfig) ([]*calendar.Event, error) {
	const maxResults = 4096
	eventsCall := c.svc.Events.List(config.CalendarID).
		Context(ctx).
		TimeMin(config.StartTime.Format(time.RFC3339)).
		TimeMax(config.EndTime.Format(time.RFC3339)).
		MaxResults(maxResults).
		SingleEvents(true).
		OrderBy("startTime")

	events, err := eventsCall.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %w", err)
	}

	items := events.Items
	return items, nil
}
