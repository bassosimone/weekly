// main.go - main file
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/bassosimone/clip"
	"github.com/bassosimone/clip/pkg/nflag"
)

// configurable for testing
var env = clip.NewStdlibExecEnv()

func main() {
	// Define the overall suite version
	const version = "0.2.0"

	// Create the ls leaf command
	lsCmd := &clip.LeafCommand[*clip.StdlibExecEnv]{
		BriefDescriptionText: "List calendar events using Calendar API.",
		RunFunc:              lsMain,
	}

	// Create the init leaf command
	initCmd := &clip.LeafCommand[*clip.StdlibExecEnv]{
		BriefDescriptionText: "Login with Calendar API and select the calendar to use.",
		RunFunc:              initMain,
	}

	// Create the root command
	rootCmd := &clip.RootCommand[*clip.StdlibExecEnv]{
		Command: &clip.DispatcherCommand[*clip.StdlibExecEnv]{
			BriefDescriptionText: "Track personal activity using Google Calendar.",
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
