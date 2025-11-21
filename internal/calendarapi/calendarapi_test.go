// calendarapi_test.go - tests for the calendarapi package
// SPDX-License-Identifier: GPL-3.0-or-later

package calendarapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func TestNewClient(t *testing.T) {

	// describes a testCase to run
	type testCase struct {
		// name of the test case
		name string

		// credentialsPath is the path to the credentials
		credentialsPath string

		// OPTIONAL function used to instantiate a new calendar instance
		calendarNewService func(ctx context.Context, opts ...option.ClientOption) (*calendar.Service, error)

		// expected valid client
		client bool

		// expected error
		err error
	}

	// test cases run by this function
	cases := []testCase{
		{
			name:            "with invalid credentials path",
			credentialsPath: filepath.Join("testdata", "nonexistent.json"),
			client:          false,
			err:             errors.New(`unable to read credentials file: open testdata/nonexistent.json: no such file or directory`),
		},

		{
			name:            "with credentials path pointing to an empty dictionary",
			credentialsPath: filepath.Join("testdata", "empty-dict.json"),
			client:          false,
			err:             errors.New(`unable to create JWT config: google: read JWT from JSON credentials: 'type' field is "" (expected "service_account")`),
		},

		{
			name:            "with failure to instantiate new service",
			credentialsPath: filepath.Join("testdata", "service-account.json"),
			calendarNewService: func(ctx context.Context, opts ...option.ClientOption) (*calendar.Service, error) {
				return nil, errors.New("mocked error")
			},
			client: false,
			err:    errors.New("unable to create calendar service: mocked error"),
		},

		{
			name:            "with success",
			credentialsPath: filepath.Join("testdata", "service-account.json"),
			client:          true,
			err:             nil,
		},
	}

	// run all test cases
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// create a background context
			ctx := context.Background()

			// possibly deal with modifying the calendar service factory
			if tc.calendarNewService != nil {
				calendarNewServiceFunc = tc.calendarNewService
			} else {
				calendarNewServiceFunc = calendar.NewService
			}

			// run the function we're testing
			client, err := NewClient(ctx, tc.credentialsPath)

			// make sure the error is the one we actually expect
			switch {
			case err == nil && tc.err == nil:
				// nothing

			case err == nil && tc.err != nil:
				t.Error("expected", tc.err, "got", err)

			case err != nil && tc.err == nil:
				t.Error("expected", tc.err, "got", err)

			case err != nil && tc.err != nil && err.Error() == tc.err.Error():
				// nothing

			default:
				t.Error("expected", tc.err, "got", err)
			}

			// make sure the client is the one we actually expect
			gotclient := client != nil
			if diff := cmp.Diff(tc.client, gotclient); diff != "" {
				t.Error(diff)
			}
		})
	}
}

// roundTripper implements [http.RoundTripper].
type roundTripper func(req *http.Request) (*http.Response, error)

var _ http.RoundTripper = roundTripper(nil)

// RoundTrip implements [http.RoundTripper].
func (fx roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return fx(req)
}

// newJSONResponse constructs a minimal *http.Response with JSON body.
func newJSONResponse(req *http.Request, status int, body string) *http.Response {
	resp := &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp
}

func TestClient_FetchEvents(t *testing.T) {

	// describes a testCase to run
	type testCase struct {
		// name of the test case
		name string

		// credentialsPath is the path to the credentials
		credentialsPath string

		// function to mock the actual round trip
		roundTrip func(req *http.Request) (*http.Response, error)

		// expected output events
		events []Event

		// expected error
		err error
	}

	// test cases run by this function
	cases := []testCase{
		{
			name:            "failure getting token",
			credentialsPath: filepath.Join("testdata", "service-account.json"),
			roundTrip: func(req *http.Request) (*http.Response, error) {
				if req.URL.Host == "oauth2.googleapis.com" && req.URL.Path == "/token" {
					return nil, errors.New("mocked token error")
				}
				return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
			},
			events: nil,
			err:    errors.New(`unable to retrieve events: Get "https://www.googleapis.com/calendar/v3/calendars//events?alt=json&maxResults=0&orderBy=startTime&prettyPrint=false&singleEvents=true&timeMax=0001-01-01T00%3A00%3A00Z&timeMin=0001-01-01T00%3A00%3A00Z": oauth2: cannot fetch token: Post "https://oauth2.googleapis.com/token": mocked token error`),
		},

		{
			name:            "failure fetching events after successful token",
			credentialsPath: filepath.Join("testdata", "service-account.json"),
			roundTrip: func(req *http.Request) (*http.Response, error) {
				switch {
				// First call: token acquisition succeeds
				case req.URL.Host == "oauth2.googleapis.com" && req.URL.Path == "/token":
					body := `{"access_token":"ya29.testtoken","token_type":"Bearer","expires_in":3600}`
					return newJSONResponse(req, http.StatusOK, body), nil

				// Second call: calendar events fails
				case req.URL.Host == "www.googleapis.com" && strings.HasPrefix(req.URL.Path, "/calendar/v3/calendars/"):
					return nil, errors.New("mocked events error")

				default:
					return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
				}
			},
			events: nil,
			err:    errors.New(`unable to retrieve events: Get "https://www.googleapis.com/calendar/v3/calendars//events?alt=json&maxResults=0&orderBy=startTime&prettyPrint=false&singleEvents=true&timeMax=0001-01-01T00%3A00%3A00Z&timeMin=0001-01-01T00%3A00%3A00Z": mocked events error`),
		},

		{
			name:            "success",
			credentialsPath: filepath.Join("testdata", "service-account.json"),
			roundTrip: func(req *http.Request) (*http.Response, error) {
				switch {
				// Token endpoint
				case req.URL.Host == "oauth2.googleapis.com" && req.URL.Path == "/token":
					body := `{"access_token":"ya29.testtoken","token_type":"Bearer","expires_in":3600}`
					return newJSONResponse(req, http.StatusOK, body), nil

				// Calendar events endpoint
				case req.URL.Host == "www.googleapis.com" && strings.HasPrefix(req.URL.Path, "/calendar/v3/calendars/"):
					body := `{
						"items": [
							{
								"summary": "Test event",
								"start": { "dateTime": "2025-01-01T10:00:00Z" },
								"end":   { "dateTime": "2025-01-01T11:00:00Z" }
							}
						]
					}`
					return newJSONResponse(req, http.StatusOK, body), nil

				default:
					return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
				}
			},
			events: []Event{
				{
					Summary:   "Test event",
					StartTime: "2025-01-01T10:00:00Z",
					EndTime:   "2025-01-01T11:00:00Z",
				},
			},
			err: nil,
		},
	}

	// run all test cases
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// create a background context with mocked HTTP client
			ctx := context.Background()
			httpClient := &http.Client{Transport: roundTripper(tc.roundTrip)}
			ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

			// instantiate the new client
			client, err := NewClient(ctx, tc.credentialsPath)
			assert.NoError(t, err)

			// create configuration for fetching events
			config := FetchEventsConfig{
				CalendarID: "",
				StartTime:  time.Time{},
				EndTime:    time.Time{},
				MaxEvents:  0,
			}

			// run the function we're testing
			events, err := client.FetchEvents(ctx, &config)

			// make sure the error is the one we actually expect
			switch {
			case err == nil && tc.err == nil:
				// nothing

			case err == nil && tc.err != nil:
				t.Error("expected", tc.err, "got", err)

			case err != nil && tc.err == nil:
				t.Error("expected", tc.err, "got", err)

			case err != nil && tc.err != nil && err.Error() == tc.err.Error():
				// nothing

			default:
				t.Error("expected", tc.err, "got", err)
			}

			// make sure the events are the one we expect
			if diff := cmp.Diff(tc.events, events); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestEvent_String(t *testing.T) {
	ev := Event{
		Summary:   "$nexa %development",
		StartTime: "2017-11-03T11:30:00+01:00",
		EndTime:   "2017-11-03T12:00:00+01:00",
	}
	expect := `{"Summary":"$nexa %development","StartTime":"2017-11-03T11:30:00+01:00","EndTime":"2017-11-03T12:00:00+01:00"}`
	if diff := cmp.Diff(expect, ev.String()); diff != "" {
		t.Fatal(diff)
	}
}
