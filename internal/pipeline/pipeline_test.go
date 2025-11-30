// pipeline_test.go - pipeline tests
// SPDX-License-Identifier: GPL-3.0-or-later

package pipeline

import (
	"errors"
	"testing"
	"time"

	"github.com/bassosimone/weekly/internal/parser"
	"github.com/google/go-cmp/cmp"
)

func mustParseTime(t *testing.T, value string) time.Time {
	tv, err := time.Parse("2006-01-02T15:04:05-07:00", value)
	if err != nil {
		t.Fatal(err)
	}
	return tv
}

func TestRun(t *testing.T) {

	// defines a test case within this function
	type testCase struct {
		// name is the test case name
		name string

		// config contains the pipeline configuration
		config *Config

		// inputs contains the input events
		inputs []parser.Event

		// outputs contains the expected output events
		outputs []parser.Event

		// err contains the expected error
		err error
	}

	// defines all the test cases
	cases := []testCase{
		{
			name:    "with empty input and empty config",
			config:  &Config{},
			inputs:  []parser.Event{},
			outputs: nil,
			err:     nil,
		},

		{
			name:   "with no filtering, aggregation, or totaling",
			config: &Config{},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			err: nil,
		},

		{
			name: "filter by single project",
			config: &Config{
				Project: "nexa",
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T14:00:00+01:00"),
					Duration:  45 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T14:00:00+01:00"),
					Duration:  45 * time.Minute,
				},
			},
			err: nil,
		},

		{
			name: "filter by single tag",
			config: &Config{
				Tag: "neubot",
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"ndt"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T14:00:00+01:00"),
					Duration:  45 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
			},
			err: nil,
		},

		{
			name: "filter by project with no matches",
			config: &Config{
				Project: "nonexistent",
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
			},
			outputs: nil,
			err:     nil,
		},

		{
			name: "aggregate daily",
			config: &Config{
				Aggregate: "daily",
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T14:00:00+01:00"),
					Duration:  45 * time.Minute,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "mlab",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-03T00:00:00+00:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-03T00:00:00+00:00"),
					Duration:  time.Hour + 45*time.Minute,
				},
			},
			err: nil,
		},

		{
			name: "aggregate daily across multiple days",
			config: &Config{
				Aggregate: "daily",
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-04T14:00:00+01:00"),
					Duration:  2 * time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-04T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-03T00:00:00+00:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-04T00:00:00+00:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-04T00:00:00+00:00"),
					Duration:  2 * time.Hour,
				},
			},
			err: nil,
		},

		{
			name: "aggregate weekly",
			config: &Config{
				Aggregate: "weekly",
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-21T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-23T14:00:00+01:00"),
					Duration:  5 * time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-24T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "mlab",
					Activity:  "development",
					Tags:      []string{"iqb"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-25T11:30:00+01:00"),
					Duration:  4 * time.Hour,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "mlab",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-20T00:00:00+00:00"),
					Duration:  4*time.Hour + 30*time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-20T00:00:00+00:00"),
					Duration:  6 * time.Hour,
				},
			},
			err: nil,
		},

		{
			name: "aggregate monthly",
			config: &Config{
				Aggregate: "monthly",
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-15T14:00:00+01:00"),
					Duration:  2 * time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-20T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "mlab",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-01T00:00:00+00:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-01T00:00:00+00:00"),
					Duration:  3 * time.Hour,
				},
			},
			err: nil,
		},

		{
			name: "aggregate monthly across multiple months",
			config: &Config{
				Aggregate: "monthly",
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-10-15T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-15T14:00:00+01:00"),
					Duration:  2 * time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-20T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-10-01T00:00:00+00:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-01T00:00:00+00:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-01T00:00:00+00:00"),
					Duration:  2 * time.Hour,
				},
			},
			err: nil,
		},

		{
			name: "aggregate with invalid policy",
			config: &Config{
				Aggregate: "invalid",
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
			},
			outputs: nil,
			err:     errors.New("invalid aggregation policy: invalid (valid values: daily, monthly)"),
		},

		{
			name: "compute total by project",
			config: &Config{
				Total: true,
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T14:00:00+01:00"),
					Duration:  45 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "mlab",
					Activity:  "",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour + 45*time.Minute,
				},
			},
			err: nil,
		},

		{
			name: "compute total with single project",
			config: &Config{
				Total: true,
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "nexa",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  90 * time.Minute,
				},
			},
			err: nil,
		},

		{
			name: "filter by project and compute total",
			config: &Config{
				Project: "nexa",
				Total:   true,
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T14:00:00+01:00"),
					Duration:  45 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour + 45*time.Minute,
				},
			},
			err: nil,
		},

		{
			name: "filter by project and aggregate daily",
			config: &Config{
				Project:   "nexa",
				Aggregate: "daily",
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-04T14:00:00+01:00"),
					Duration:  2 * time.Hour,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-03T00:00:00+00:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      nil,
					Persons:   nil,
					StartTime: mustParseTime(t, "2017-11-04T00:00:00+00:00"),
					Duration:  2 * time.Hour,
				},
			},
			err: nil,
		},

		{
			name: "aggregate daily and compute total",
			config: &Config{
				Aggregate: "daily",
				Total:     true,
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-04T14:00:00+01:00"),
					Duration:  2 * time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "mlab",
					Activity:  "",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T00:00:00+00:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T00:00:00+00:00"),
					Duration:  3 * time.Hour,
				},
			},
			err: nil,
		},

		{
			name: "aggregate monthly and compute total",
			config: &Config{
				Aggregate: "monthly",
				Total:     true,
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-10-15T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-15T14:00:00+01:00"),
					Duration:  2 * time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-20T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "mlab",
					Activity:  "",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-01T00:00:00+00:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-10-01T00:00:00+00:00"),
					Duration:  3 * time.Hour,
				},
			},
			err: nil,
		},

		{
			name: "all features combined: filter, aggregate, and total",
			config: &Config{
				Project:   "nexa",
				Aggregate: "daily",
				Total:     true,
			},
			inputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-04T14:00:00+01:00"),
					Duration:  2 * time.Hour,
				},
				{
					Project:   "mlab",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"alice"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			outputs: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T00:00:00+00:00"),
					Duration:  3 * time.Hour,
				},
			},
			err: nil,
		},

		{
			name: "aggregate empty input",
			config: &Config{
				Aggregate: "daily",
			},
			inputs:  []parser.Event{},
			outputs: nil,
			err:     nil,
		},

		{
			name: "total with empty input",
			config: &Config{
				Total: true,
			},
			inputs:  []parser.Event{},
			outputs: []parser.Event{},
			err:     nil,
		},
	}

	// runs each test case in sequence
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// invoke the function that we're testing
			outputs, err := Run(tc.config, tc.inputs)

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

			// compare the outputs
			if diff := cmp.Diff(tc.outputs, outputs); diff != "" {
				t.Error(diff)
			}
		})
	}
}
