// calendarapi.go - use the Google Calendar API
// SPDX-License-Identifier: GPL-3.0-or-later

// Package calendarapi allows using the Google Calendar API.
package calendarapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Client is a Google Calendar API client.
//
// The zero value is invalid: construct with the [NewClient] factory.
type Client struct {
	svc *calendar.Service
}

// Allows overriding [calendar.NewService] in the test suite.
var calendarNewServiceFunc = calendar.NewService

// NewClient creates a new Calendar API client using service account credentials.
//
// The ctx argument allows to cancel a pending call.
//
// The credentialsPath argument is the file path containing the service account credentials.
//
// The return value is either a valid [*Client] or an error.
func NewClient(ctx context.Context, credentialsPath string) (*Client, error) {
	// Read the service account credentials
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w", err)
	}

	// This function uses the private key in the JSON file to create a JWT,
	// which is used by the service-account authentication flow.
	//
	// We use the CalendarReadonlyScope for security (least privilege).
	config, err := google.JWTConfigFromJSON(data, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to create JWT config: %w", err)
	}

	// The JWT config handles the authentication process automatically:
	//
	// 1. Signs the JWT with the private key.
	//
	// 2. Exchanges the JWT for an access token with Google's auth server.
	//
	// 3. Automatically refreshes the access token when it expires.
	httpClient := config.Client(ctx)

	// Create the calendar service
	service, err := calendarNewServiceFunc(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create calendar service: %w", err)
	}

	return &Client{svc: service}, nil
}

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

	// MaxEvents is the MANDATORY number of maximum events to fetch.
	MaxEvents int64
}

// Event is the kind of event emitted by this package.
//
// It simplifies the actually-fetched event by removing unnecessary fields
// and making the result straightforward to parse.
type Event struct {
	// Summary is the calendar event summary
	Summary string

	// StartTime is the calendar event start time
	StartTime string

	// EndTime is the calendar event end time
	EndTime string
}

func newEventList(inputs []*calendar.Event) (outputs []Event) {
	for _, ev := range inputs {
		outputs = append(outputs, newEvent(ev))
	}
	return
}

func newEvent(ev *calendar.Event) Event {
	return Event{
		Summary:   ev.Summary,
		StartTime: ev.Start.DateTime,
		EndTime:   ev.End.DateTime,
	}
}

func (ev *Event) String() string {
	// Note: json.Marshal cannot fail for this structure
	data, _ := json.Marshal(ev)
	return string(data)
}

// FetchEvents retrieves calendar events within the specified time range.
//
// The ctx argument allows to cancel a pending call.
//
// The calendarID argument is the string identifier of the calendar.
//
// The timeMin, timeMax arguments identify the time range.
//
// The return value is either a non-empty slice of [Event] or an error.
func (c *Client) FetchEvents(ctx context.Context, config *FetchEventsConfig) ([]Event, error) {
	eventsCall := c.svc.Events.List(config.CalendarID).
		Context(ctx).
		TimeMin(config.StartTime.Format(time.RFC3339)).
		TimeMax(config.EndTime.Format(time.RFC3339)).
		MaxResults(config.MaxEvents).
		SingleEvents(true).
		OrderBy("startTime")

	events, err := eventsCall.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %w", err)
	}

	items := newEventList(events.Items)
	return items, nil
}
