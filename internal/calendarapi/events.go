// events.go - code to fetch events
// SPDX-License-Identifier: GPL-3.0-or-later

package calendarapi

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
)

// FetchEvents retrieves calendar events within the specified time range.
//
// The ctx argument allows to cancel a pending call.
//
// The calendarID argument is the string identifier of the calendar.
//
// The timeMin, timeMax arguments identify the time range.
//
// The return value is either a non-empty list or an error.
func (c *Client) FetchEvents(
	ctx context.Context, calendarID string, timeMin, timeMax time.Time) ([]*calendar.Event, error) {
	const maxResults = 4096
	eventsCall := c.svc.Events.List(calendarID).
		Context(ctx).
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
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
