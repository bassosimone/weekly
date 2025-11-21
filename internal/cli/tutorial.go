// tutorial.go - tutorial subcommand
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/bassosimone/clip"
	"github.com/bassosimone/clip/pkg/assert"
	"github.com/bassosimone/clip/pkg/nflag"
)

//go:embed tutorial.md
var tutorialData string

// tutorialMain is the main entry point for the `tutorial` leaf command.
func tutorialMain(ctx context.Context, args *clip.CommandArgs[*clip.StdlibExecEnv]) error {
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

	// Parse the flags
	assert.NotError(fset.Parse(args.Args))

	// Print the tutorial data to stdout
	fmt.Printf("%s", tutorialData)
	return nil
}
