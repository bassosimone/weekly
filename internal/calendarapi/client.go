// client.go - Client implementation
// SPDX-License-Identifier: GPL-3.0-or-later

package calendarapi

import (
	"context"
	"fmt"
	"os"

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

// NewClient creates a new Calendar API client using OAuth2 credentials.
//
// The ctx argument allows to cancel a pending call.
//
// The credentialsPath argument is the path to the credentials file.
//
// The tokenPath argument is the path to the token file.
//
// The return value is either a valid [*Client] or an error.
func NewClient(ctx context.Context, credentialsPath, tokenPath string) (*Client, error) {
	// Read the credentials that identify the app on Google Cloud
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w", err)
	}

	// TODO(bassosimone): document that I have created a weekly-client-2025-11-18
	// application identifier, have marked it as testing, and have assigned myself
	// as the only person who is allowed to use this client.

	// Parse credentials and create OAuth2 config
	config, err := google.ConfigFromJSON(data, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %w", err)
	}

	// Get OAuth2 token either from file or via auth flow
	token, err := getToken(ctx, config, tokenPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get token: %w", err)
	}

	// Create the calendar service
	httpClient := config.Client(ctx, token)
	service, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create calendar service: %w", err)
	}

	return &Client{svc: service}, nil
}
