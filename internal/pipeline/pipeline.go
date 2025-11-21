// pipeline.go - pipeline for processing events
// SPDX-License-Identifier: GPL-3.0-or-later

// Package pipeline defines the pipeline for processing events
package pipeline

import (
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/bassosimone/weekly/internal/parser"
)

// Config contains the pipeline config.
type Config struct {
	// Aggregate OPTIONALLY aggregates events by project and policy.
	//
	// Valid policies are: monthly and weekly.
	Aggregate string

	// Project is the OPTIONAL project to filter the events for.
	Project string

	// Total OPTIONALLY sums the total time by project.
	Total bool
}

// Run runs the pipeline and returns a subset of the original events.
func Run(config *Config, events []parser.Event) ([]parser.Event, error) {
	// Maybe filter events by project
	events = maybeFilterByProject(config.Project, events)

	// Maybe create daily or monthly aggregates
	events, err := maybeAggregate(config.Aggregate, events)
	if err != nil {
		return nil, err
	}

	// Maybe sum time spent by project
	events = maybeComputeTotal(config.Total, events)

	return events, nil
}

func maybeFilterByProject(project string, inputs []parser.Event) (outputs []parser.Event) {
	for _, ev := range inputs {
		if project == "" || ev.Project == project {
			outputs = append(outputs, ev)
		}
	}
	return
}

func maybeAggregate(policy string, inputs []parser.Event) (outputs []parser.Event, err error) {
	// Honor the policy
	var timeFormat string
	switch policy {
	case "":
		return inputs, nil
	case "daily":
		timeFormat = "2006-01-02"
	case "monthly":
		timeFormat = "2006-01"
	default:
		return nil, fmt.Errorf("invalid aggregation policy: %s (valid values: daily, monthly)", policy)
	}

	// Aggregate by time period, project
	sums := make(map[string]map[string]time.Duration)
	for _, ev := range inputs {
		timeKey := ev.StartTime.Format(timeFormat)
		if sums[timeKey] == nil {
			sums[timeKey] = make(map[string]time.Duration)
		}
		sums[timeKey][ev.Project] += ev.Duration
	}

	// Generate aggregate output slice
	for _, timeKey := range slices.Sorted(maps.Keys(sums)) {
		// Note that the format must be correct since we serialized it above
		day, _ := time.Parse(timeFormat, timeKey)
		for _, project := range slices.Sorted(maps.Keys(sums[timeKey])) {
			duration := sums[timeKey][project]
			outputs = append(outputs, parser.Event{
				Project:   project,
				StartTime: day,
				Duration:  duration,
			})
		}
	}
	return
}

func maybeComputeTotal(total bool, inputs []parser.Event) []parser.Event {
	switch total {
	case true:
		sum := make(map[string]*parser.Event)
		for _, ev := range inputs {
			if _, ok := sum[ev.Project]; !ok {
				sum[ev.Project] = &parser.Event{
					Project:   ev.Project,
					Activity:  "",
					Tags:      []string{},
					Persons:   []string{},
					StartTime: ev.StartTime,
					Duration:  ev.Duration,
				}
				continue
			}
			sum[ev.Project].Duration += ev.Duration
		}

		outputs := make([]parser.Event, 0, len(sum))
		for _, key := range slices.Sorted(maps.Keys(sum)) {
			outputs = append(outputs, *sum[key])
		}
		return outputs

	default:
		return inputs
	}
}
