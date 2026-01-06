// tutorial.go - tutorial subcommand
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"context"
	_ "embed"

	"github.com/bassosimone/must"
	"github.com/bassosimone/runtimex"
	"github.com/bassosimone/vflag"
)

//go:embed tutorial.md
var tutorialData string

// tutorialBriefDescription is the `tutorial` leaf command brief description.
const tutorialBriefDescription = "Show detailed tutorial explaining the tool usage."

// tutorialMain is the main entry point for the `tutorial` leaf command.
func tutorialMain(ctx context.Context, args []string) error {
	// Create flag set
	fset := vflag.NewFlagSet("weekly tutorial", vflag.ExitOnError)
	usage := vflag.NewDefaultUsagePrinter()
	usage.AddDescription(tutorialBriefDescription)
	fset.UsagePrinter = usage

	// Not strictly needed in production but necessary for testing
	fset.Exit = env.Exit
	fset.Stderr = env.Stderr
	fset.Stdout = env.Stdout

	// Add the --help flag
	fset.AutoHelp('h', "help", "Print this help message and exit.")

	// Parse the flags
	runtimex.PanicOnError0(fset.Parse(args))

	// Print the tutorial data to stdout
	must.Fprintf(env.Stdout, "%s", tutorialData)
	return nil
}
