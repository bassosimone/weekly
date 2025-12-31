// init.go - init subcommand
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"context"
	"fmt"

	"github.com/bassosimone/runtimex"
	"github.com/bassosimone/vflag"
)

// initBriefDescription is the `init` leaf command brief description.
const initBriefDescription = "Initialize and select the calendar to use."

// initMain is the main entry point for the `init` leaf command.
func initMain(ctx context.Context, args []string) error {
	// Create flag set
	fset := vflag.NewFlagSet("weekly init", vflag.ExitOnError)
	fset.AddDescription(initBriefDescription)

	// Not strictly needed in production but necessary for testing
	fset.Exit = env.Exit
	fset.Stderr = env.Stderr
	fset.Stdout = env.Stdout

	// Create default values for flags
	var (
		configDir = xdgConfigHome(env)
	)

	// Add the --config-dir flag
	fset.StringVar(&configDir, 0, "config-dir", "Directory containing the configuration.")

	// Add the --help flag
	fset.AutoHelp('h', "help", "Print this help message and exit.")

	// Parse the flags
	runtimex.PanicOnError0(fset.Parse(args))

	// Read the calendar ID
	var cinfo calendarInfo
	fmt.Fprintf(env.Stdout, "Please, provide the default calendar ID: ")
	_ = runtimex.LogFatalOnError1(fmt.Fscanf(env.Stdin, "%s", &cinfo.ID))

	// Write the calendar ID
	runtimex.LogFatalOnError0(writeCalendarInfo(env, calendarPath(configDir), &cinfo))
	return nil
}
