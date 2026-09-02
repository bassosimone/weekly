// ls_test.go - ls subcommand tests
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// testLocation is the fixed time zone used by these tests, so that
// their results do not depend on the machine running them.
var testLocation = time.FixedZone("CET", 3600)

func Test_lsParseAsOf(t *testing.T) {
	// now is the current time as seen by the tests
	now := time.Date(2026, 9, 2, 15, 4, 5, 0, testLocation)

	type testCase struct {
		// name is the name of this test case
		name string

		// value is the raw `--as-of` flag value
		value string

		// expect is the expected anchor day
		expect time.Time

		// expectErr is the expected error string, if any
		expectErr string
	}

	cases := []testCase{{
		name:   "the empty value selects the current time",
		value:  "",
		expect: now,
	}, {
		name:   "a valid date is parsed in the location of now",
		value:  "2026-08-31",
		expect: time.Date(2026, 8, 31, 0, 0, 0, 0, testLocation),
	}, {
		name:      "a date using another format is rejected",
		value:     "31/08/2026",
		expectErr: `invalid --as-of value: parsing time "31/08/2026" as "2006-01-02": cannot parse "31/08/2026" as "2006"`,
	}, {
		name:      "a nonexisting day is rejected",
		value:     "2026-02-30",
		expectErr: `invalid --as-of value: parsing time "2026-02-30": day out of range`,
	}, {
		name:      "a date carrying a time of day is rejected",
		value:     "2026-08-31T10:00:00Z",
		expectErr: `invalid --as-of value: parsing time "2026-08-31T10:00:00Z": extra text: "T10:00:00Z"`,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := lsParseAsOf(tc.value, now)

			if tc.expectErr != "" {
				assert.EqualError(t, err, tc.expectErr)
				assert.True(t, got.IsZero())
				return
			}

			assert.NoError(t, err)
			assert.True(t, got.Equal(tc.expect), "got %v, expected %v", got, tc.expect)
			assert.Equal(t, tc.expect.Location(), got.Location())
		})
	}
}

func Test_lsDaysToTimeInterval(t *testing.T) {
	type testCase struct {
		// name is the name of this test case
		name string

		// anchor is the most recent day to fetch
		anchor time.Time

		// days is the number of days in the past to fetch
		days int64

		// expectStart is the expected (included) interval start
		expectStart time.Time

		// expectEnd is the expected (excluded) interval end
		expectEnd time.Time
	}

	cases := []testCase{{
		name:        "a single day selects the whole anchor day",
		anchor:      time.Date(2026, 9, 2, 15, 4, 5, 0, testLocation),
		days:        1,
		expectStart: time.Date(2026, 9, 2, 0, 0, 0, 0, testLocation),
		expectEnd:   time.Date(2026, 9, 3, 0, 0, 0, 0, testLocation),
	}, {
		name:        "the last day of a month plus its length selects the whole month",
		anchor:      time.Date(2026, 8, 31, 0, 0, 0, 0, testLocation),
		days:        31,
		expectStart: time.Date(2026, 8, 1, 0, 0, 0, 0, testLocation),
		expectEnd:   time.Date(2026, 9, 1, 0, 0, 0, 0, testLocation),
	}, {
		name:        "zero days selects the empty interval after the anchor day",
		anchor:      time.Date(2026, 9, 2, 15, 4, 5, 0, testLocation),
		days:        0,
		expectStart: time.Date(2026, 9, 3, 0, 0, 0, 0, testLocation),
		expectEnd:   time.Date(2026, 9, 3, 0, 0, 0, 0, testLocation),
	}, {
		name:        "a negative number of days is clamped to zero",
		anchor:      time.Date(2026, 9, 2, 15, 4, 5, 0, testLocation),
		days:        -128,
		expectStart: time.Date(2026, 9, 3, 0, 0, 0, 0, testLocation),
		expectEnd:   time.Date(2026, 9, 3, 0, 0, 0, 0, testLocation),
	}, {
		name:        "too many days are clamped to a year",
		anchor:      time.Date(2026, 9, 2, 15, 4, 5, 0, testLocation),
		days:        4096,
		expectStart: time.Date(2025, 9, 3, 0, 0, 0, 0, testLocation),
		expectEnd:   time.Date(2026, 9, 3, 0, 0, 0, 0, testLocation),
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start, end := lsDaysToTimeInterval(tc.anchor, tc.days)
			assert.True(t, start.Equal(tc.expectStart), "start: got %v, expected %v", start, tc.expectStart)
			assert.True(t, end.Equal(tc.expectEnd), "end: got %v, expected %v", end, tc.expectEnd)
		})
	}
}
