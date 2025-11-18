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
	service, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create calendar service: %w", err)
	}

	return &Client{svc: service}, nil
}
