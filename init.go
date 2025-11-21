// init.go - init subcommand
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"fmt"

	"github.com/bassosimone/clip"
	"github.com/bassosimone/clip/pkg/assert"
	"github.com/bassosimone/clip/pkg/nflag"
)

// initMain is the main entry point for the `init` leaf command.
func initMain(ctx context.Context, args *clip.CommandArgs[*clip.StdlibExecEnv]) error {
	// Create flag set
	fset := nflag.NewFlagSet(args.CommandName, nflag.ExitOnError)
	fset.Description = args.Command.BriefDescription()
	fset.PositionalArgumentsUsage = ""
	fset.MinPositionalArgs = 0
	fset.MaxPositionalArgs = 0

	// Not strictly needed in production but necessary for testing
	fset.Exit = args.Env.Exit
	fset.Stderr = args.Env.Stderr()
	fset.Stdout = args.Env.Stdout()

	// Create default values for flags
	var (
		configDir = xdgConfigHome(args.Env)
	)

	// Add the --config-dir flag
	fset.StringFlagVar(&configDir, "config-dir", 0, "Directory containing the configuration.")

	// Add the --help flag
	fset.AutoHelp("help", 'h', "Print this help message and exit.")

	// Parse the flags
	assert.NotError(fset.Parse(args.Args))

	// Read the calendar ID
	var cinfo calendarInfo
	fmt.Printf("Please, provide the default calendar ID: ")
	fmt.Scanf("%s", &cinfo.ID)

	// Write the calendar ID
	must0(writeCalendarInfo(calendarPath(configDir), &cinfo))
	return nil
}
