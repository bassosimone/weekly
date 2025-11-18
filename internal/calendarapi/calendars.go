// calendars.go - Listing calendars
// SPDX-License-Identifier: GPL-3.0-or-later

package calendarapi

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/api/calendar/v3"
)

// ListCalendars returns all available calendars for the authenticated user.
//
// The ctx argument allows to cancel a pending call.
//
// The return value is either a non-empty list or an error.
func (c *Client) ListCalendars(ctx context.Context) ([]*calendar.CalendarListEntry, error) {
	calendarList, err := c.svc.CalendarList.List().Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve calendar list: %w", err)
	}

	items := calendarList.Items
	if len(items) <= 0 {
		return nil, errors.New("no calendars returned")
	}

	return items, nil
}
