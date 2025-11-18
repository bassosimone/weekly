// main.go - main file
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"runtime/debug"

	"github.com/bassosimone/clip"
	"github.com/bassosimone/clip/pkg/nflag"
)

// configurable for testing
var env = clip.NewStdlibExecEnv()

func main() {
	// Define the overall suite version
	version := "unknown"
	if binfo, ok := debug.ReadBuildInfo(); ok {
		version = binfo.Main.Version
	}

	// Create the ls leaf command
	lsCmd := &clip.LeafCommand[*clip.StdlibExecEnv]{
		BriefDescriptionText: "List events from the selected calendar.",
		RunFunc:              lsMain,
	}

	// Create the init leaf command
	initCmd := &clip.LeafCommand[*clip.StdlibExecEnv]{
		BriefDescriptionText: "Initialize and select the calendar to use.",
		RunFunc:              initMain,
	}

	// Create the root command
	rootCmd := &clip.RootCommand[*clip.StdlibExecEnv]{
		Command: &clip.DispatcherCommand[*clip.StdlibExecEnv]{
			BriefDescriptionText: "Track weekly activity using Google Calendar.",
			Commands: map[string]clip.Command[*clip.StdlibExecEnv]{
				"ls":   lsCmd,
				"init": initCmd,
			},
			ErrorHandling:             nflag.ExitOnError,
			Version:                   version,
			OptionPrefixes:            []string{"-", "--"},
			OptionsArgumentsSeparator: "--",
		},
		AutoCancel: true,
	}

	// Execute the root command
	rootCmd.Main(env)
}
