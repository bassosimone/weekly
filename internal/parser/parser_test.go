// parser_test.go - parser tests
// SPDX-License-Identifier: GPL-3.0-or-later

package parser

import (
	"errors"
	"testing"
	"time"

	"github.com/bassosimone/weekly/internal/calendarapi"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func mustParseTime(t *testing.T, value string) time.Time {
	tv, err := time.Parse(timeFormat, value)
	assert.NoError(t, err)
	return tv
}

func TestParser(t *testing.T) {

	// defines a test case within this function
	type testCase struct {
		// name is the test case name
		name string

		// inputs contains the inputs to provide to this function
		inputs []calendarapi.Event

		// outputs contains the expected outputs
		outputs []Event

		// err contains the expected error
		err error
	}

	// defines all the test cases
	cases := []testCase{
		{
			name:    "with empty input",
			inputs:  []calendarapi.Event{},
			outputs: []Event{},
			err:     nil,
		},

		{
			name: "with valid input",
			inputs: []calendarapi.Event{
				{
					Summary:   "$nexa %development #neubot #pr42",
					StartTime: "2017-11-03T10:00:00+01:00",
					EndTime:   "2017-11-03T11:00:00+01:00",
				},
				{
					Summary:   "$nexa %meeting #staff @fmorando @alemela @riemma",
					StartTime: "2017-11-03T11:30:00+01:00",
					EndTime:   "2017-11-03T12:00:00+01:00",
				},
				{
					Summary:   "$nexa %development #neubot #pr42",
					StartTime: "2017-11-03T12:15:00+01:00",
					EndTime:   "2017-11-03T13:00:00+01:00",
				},
			},
			outputs: []Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot", "pr42"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
				{
					Project:   "nexa",
					Activity:  "meeting",
					Tags:      []string{"staff"},
					Persons:   []string{"fmorando", "alemela", "riemma"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot", "pr42"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T12:15:00+01:00"),
					Duration:  45 * time.Minute,
				},
			},
			err: nil,
		},

		{
			name: "with empty event summary",
			inputs: []calendarapi.Event{
				{
					Summary:   "",
					StartTime: "2017-11-03T11:30:00+01:00",
					EndTime:   "2017-11-03T12:00:00+01:00",
				},
			},
			outputs: nil,
			err:     errors.New(`no project or activity in {"Summary":"","StartTime":"2017-11-03T11:30:00+01:00","EndTime":"2017-11-03T12:00:00+01:00"}`),
		},

		{
			name: "with extra tokens event summary",
			inputs: []calendarapi.Event{
				{
					Summary:   "we just ignore $nexa %development extra tokens",
					StartTime: "2017-11-03T11:30:00+01:00",
					EndTime:   "2017-11-03T12:00:00+01:00",
				},
			},
			outputs: []Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			err: nil,
		},

		{
			name: "with invalid StartTime",
			inputs: []calendarapi.Event{
				{
					Summary:   "$nexa %development",
					StartTime: "invalid",
					EndTime:   "2017-11-03T12:00:00+01:00",
				},
			},
			outputs: nil,
			err:     errors.New(`invalid start time in {"Summary":"$nexa %development","StartTime":"invalid","EndTime":"2017-11-03T12:00:00+01:00"}: parsing time "invalid" as "2006-01-02T15:04:05-07:00": cannot parse "invalid" as "2006"`),
		},

		{
			name: "with invalid EndTime",
			inputs: []calendarapi.Event{
				{
					Summary:   "$nexa %development",
					StartTime: "2017-11-03T11:30:00+01:00",
					EndTime:   "invalid",
				},
			},
			outputs: nil,
			err:     errors.New(`invalid end time in {"Summary":"$nexa %development","StartTime":"2017-11-03T11:30:00+01:00","EndTime":"invalid"}: parsing time "invalid" as "2006-01-02T15:04:05-07:00": cannot parse "invalid" as "2006"`),
		},

		{
			name: "with multiple projects",
			inputs: []calendarapi.Event{
				{
					Summary:   "$nexa $development",
					StartTime: "2017-11-03T11:30:00+01:00",
					EndTime:   "2017-11-03T12:00:00+01:00",
				},
			},
			outputs: nil,
			err:     errors.New(`multiple projects in {"Summary":"$nexa $development","StartTime":"2017-11-03T11:30:00+01:00","EndTime":"2017-11-03T12:00:00+01:00"}`),
		},

		{
			name: "with multiple activities",
			inputs: []calendarapi.Event{
				{
					Summary:   "%nexa %development",
					StartTime: "2017-11-03T11:30:00+01:00",
					EndTime:   "2017-11-03T12:00:00+01:00",
				},
			},
			outputs: nil,
			err:     errors.New(`multiple activities in {"Summary":"%nexa %development","StartTime":"2017-11-03T11:30:00+01:00","EndTime":"2017-11-03T12:00:00+01:00"}`),
		},
	}

	// runs each test case in sequence
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// invoke the function that we're testing
			outputs, err := Parse(tc.inputs)

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
