// output_test.go - output tests
// SPDX-License-Identifier: GPL-3.0-or-later

package output

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"strings"
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

func TestWrite(t *testing.T) {

	// defines a test case within this function
	type testCase struct {
		// name is the test case name
		name string

		// format is the output format
		format string

		// events contains the input events
		events []parser.Event

		// expected contains the expected output string
		expected string

		// err contains the expected error
		err error

		// skipExactMatch if true, only checks that output is not empty
		skipExactMatch bool
	}

	// defines all the test cases
	cases := []testCase{
		{
			name:     "with invalid format",
			format:   "invalid",
			events:   []parser.Event{},
			expected: "",
			err:      errors.New("the --format flag accepts one of these values: box, csv, invoice, json"),
		},

		{
			name:     "json with empty input",
			format:   "json",
			events:   []parser.Event{},
			expected: "",
			err:      nil,
		},

		{
			name:   "json with single event",
			format: "json",
			events: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot", "pr42"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
			},
			expected: `{"Project":"nexa","Activity":"development","Tags":["neubot","pr42"],"Persons":[],"StartTime":"2017-11-03T10:00:00+01:00","Duration":3600000000000}` + "\n",
			err:      nil,
		},

		{
			name:   "json with multiple events",
			format: "json",
			events: []parser.Event{
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
					Persons:   []string{"alice", "bob"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			expected: `{"Project":"nexa","Activity":"development","Tags":["neubot"],"Persons":[],"StartTime":"2017-11-03T10:00:00+01:00","Duration":3600000000000}` + "\n" +
				`{"Project":"mlab","Activity":"meeting","Tags":["staff"],"Persons":["alice","bob"],"StartTime":"2017-11-03T11:30:00+01:00","Duration":1800000000000}` + "\n",
			err: nil,
		},

		{
			name:   "json with empty tags and persons",
			format: "json",
			events: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
			},
			expected: `{"Project":"nexa","Activity":"development","Tags":[],"Persons":[],"StartTime":"2017-11-03T10:00:00+01:00","Duration":3600000000000}` + "\n",
			err:      nil,
		},

		{
			name:     "csv with empty input",
			format:   "csv",
			events:   []parser.Event{},
			expected: "",
			err:      nil,
		},

		{
			name:   "csv with single event",
			format: "csv",
			events: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot", "pr42"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
			},
			expected: "2017-11-03T10:00:00+01:00,1h0m0s,nexa,development,neubot pr42,\n",
			err:      nil,
		},

		{
			name:   "csv with multiple events",
			format: "csv",
			events: []parser.Event{
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
					Persons:   []string{"alice", "bob"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			expected: "2017-11-03T10:00:00+01:00,1h0m0s,nexa,development,neubot,\n" +
				"2017-11-03T11:30:00+01:00,30m0s,mlab,meeting,staff,alice bob\n",
			err: nil,
		},

		{
			name:   "csv with empty tags and persons",
			format: "csv",
			events: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
			},
			expected: "2017-11-03T10:00:00+01:00,1h0m0s,nexa,development,,\n",
			err:      nil,
		},

		{
			name:     "invoice with empty input",
			format:   "invoice",
			events:   []parser.Event{},
			expected: "",
			err:      nil,
		},

		{
			name:   "invoice with single event",
			format: "invoice",
			events: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
			},
			expected: "nexa,2017-11-03,1\n",
			err:      nil,
		},

		{
			name:   "invoice with multiple events",
			format: "invoice",
			events: []parser.Event{
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
			expected: "nexa,2017-11-03,1\nmlab,2017-11-03,0.5\n",
			err:      nil,
		},

		{
			name:   "invoice with fractional hours",
			format: "invoice",
			events: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  45 * time.Minute,
				},
			},
			expected: "nexa,2017-11-03,0.75\n",
			err:      nil,
		},

		{
			name:   "box with empty input",
			format: "box",
			events: []parser.Event{},
			// Box format will still produce a header
			skipExactMatch: true,
			err:            nil,
		},

		{
			name:   "box with single event",
			format: "box",
			events: []parser.Event{
				{
					Project:   "nexa",
					Activity:  "development",
					Tags:      []string{"neubot"},
					Persons:   []string{},
					StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
					Duration:  time.Hour,
				},
			},
			// Box format produces ASCII table, just verify it's not empty
			skipExactMatch: true,
			err:            nil,
		},

		{
			name:   "box with multiple events",
			format: "box",
			events: []parser.Event{
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
					Persons:   []string{"alice", "bob"},
					StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
					Duration:  30 * time.Minute,
				},
			},
			skipExactMatch: true,
			err:            nil,
		},
	}

	// runs each test case in sequence
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// create a buffer to capture output
			var buf bytes.Buffer

			// invoke the function that we're testing
			err := Write(&buf, tc.format, tc.events)

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
			output := buf.String()
			if tc.skipExactMatch {
				// For box format, just verify we got some output
				if len(tc.events) > 0 && output == "" {
					t.Error("expected non-empty output for box format")
				}
			} else {
				if diff := cmp.Diff(tc.expected, output); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

// TestWriteJSON_Roundtrip tests that JSON output can be unmarshaled back
func TestWriteJSON_Roundtrip(t *testing.T) {
	events := []parser.Event{
		{
			Project:   "nexa",
			Activity:  "development",
			Tags:      []string{"neubot", "pr42"},
			Persons:   []string{"alice"},
			StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
			Duration:  time.Hour,
		},
		{
			Project:   "mlab",
			Activity:  "meeting",
			Tags:      []string{},
			Persons:   []string{},
			StartTime: mustParseTime(t, "2017-11-03T11:30:00+01:00"),
			Duration:  30 * time.Minute,
		},
	}

	var buf bytes.Buffer
	err := Write(&buf, "json", events)
	if err != nil {
		t.Fatal(err)
	}

	// Parse each line back
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != len(events) {
		t.Fatalf("expected %d lines, got %d", len(events), len(lines))
	}

	for i, line := range lines {
		var ev parser.Event
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Errorf("line %d: failed to unmarshal: %v", i, err)
			continue
		}

		if diff := cmp.Diff(events[i], ev); diff != "" {
			t.Errorf("line %d: %s", i, diff)
		}
	}
}

// TestWriteCSV_Roundtrip tests that CSV output can be parsed back
func TestWriteCSV_Roundtrip(t *testing.T) {
	events := []parser.Event{
		{
			Project:   "nexa",
			Activity:  "development",
			Tags:      []string{"neubot", "pr42"},
			Persons:   []string{"alice"},
			StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
			Duration:  time.Hour,
		},
	}

	var buf bytes.Buffer
	err := Write(&buf, "csv", events)
	if err != nil {
		t.Fatal(err)
	}

	// Parse the CSV back
	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != len(events) {
		t.Fatalf("expected %d records, got %d", len(events), len(records))
	}

	// Verify the first record
	record := records[0]
	if len(record) != 6 {
		t.Fatalf("expected 6 fields, got %d", len(record))
	}

	// Check project
	if record[2] != "nexa" {
		t.Errorf("expected project 'nexa', got %q", record[2])
	}

	// Check activity
	if record[3] != "development" {
		t.Errorf("expected activity 'development', got %q", record[3])
	}

	// Check tags
	if record[4] != "neubot pr42" {
		t.Errorf("expected tags 'neubot pr42', got %q", record[4])
	}

	// Check persons
	if record[5] != "alice" {
		t.Errorf("expected persons 'alice', got %q", record[5])
	}
}

// TestWriteInvoice_Format tests the invoice format structure
func TestWriteInvoice_Format(t *testing.T) {
	events := []parser.Event{
		{
			Project:   "nexa",
			Activity:  "development",
			Tags:      []string{},
			Persons:   []string{},
			StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
			Duration:  2*time.Hour + 30*time.Minute,
		},
	}

	var buf bytes.Buffer
	err := Write(&buf, "invoice", events)
	if err != nil {
		t.Fatal(err)
	}

	// Parse the CSV
	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	record := records[0]
	if len(record) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(record))
	}

	// Check project
	if record[0] != "nexa" {
		t.Errorf("expected project 'nexa', got %q", record[0])
	}

	// Check date format
	if record[1] != "2017-11-03" {
		t.Errorf("expected date '2017-11-03', got %q", record[1])
	}

	// Check hours (2.5 hours)
	if record[2] != "2.5" {
		t.Errorf("expected hours '2.5', got %q", record[2])
	}
}

// failingWriter is a writer that always fails
type failingWriter struct{}

func (fw *failingWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

// TestWriteJSON_WriterError tests that JSON format returns writer errors
func TestWriteJSON_WriterError(t *testing.T) {
	events := []parser.Event{
		{
			Project:   "nexa",
			Activity:  "development",
			Tags:      []string{},
			Persons:   []string{},
			StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
			Duration:  time.Hour,
		},
	}

	fw := &failingWriter{}
	err := Write(fw, "json", events)
	if err == nil {
		t.Error("expected error from failing writer, got nil")
	}
}

// TestWriteBox_ContainsExpectedData tests that box output contains expected data
func TestWriteBox_ContainsExpectedData(t *testing.T) {
	events := []parser.Event{
		{
			Project:   "nexa",
			Activity:  "development",
			Tags:      []string{"neubot"},
			Persons:   []string{"alice"},
			StartTime: mustParseTime(t, "2017-11-03T10:00:00+01:00"),
			Duration:  time.Hour,
		},
	}

	var buf bytes.Buffer
	err := Write(&buf, "box", events)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// Verify the output contains expected data
	expectedStrings := []string{
		"START TIME",
		"HOURS",
		"PROJECT",
		"ACTIVITY",
		"TAGS",
		"PERSONS",
		"nexa",
		"development",
		"neubot",
		"alice",
		"2017-11-03 10:00",
		"1.0",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, but it didn't.\nOutput:\n%s", expected, output)
		}
	}
}
