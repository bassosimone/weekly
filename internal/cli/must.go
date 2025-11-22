// must.go - must0 and must1 func
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file implements the must pattern for fatal errors in main functions.
//
// This follows the approach used by M-Lab's rtx.Must for CLI tools where
// there is no error recovery and crashing with a message is appropriate.
// The pattern keeps main function logic linear and focused on the happy path,
// rather than being dominated by repetitive error handling boilerplate.
//
// See: github.com/m-lab/go/rtx

package cli

import "fmt"

// must0 terminates the program with a fatal error message if err is non-nil.
//
// This function is used in CLI main functions where error recovery is not
// possible and the only sensible action is to exit with an error message.
func must0(env *execEnv, err error) {
	if err != nil {
		fmt.Fprintf(env.Stderr(), "fatal: %s", err.Error())
		env.Exit(1)
	}
}

// must1 is a generic version of must0 that returns a value on success.
//
// This allows writing:
//
//	client := must1(env.NewCalendarClient(ctx, path))
//
// instead of:
//
//	client, err := env.NewCalendarClient(ctx, path)
//	must0(env, err)
//
// Note: This function uses the global execEnv variable. This is a known
// tradeoff to keep the usage pattern clean.
func must1[T any](value T, err error) T {
	must0(env, err)
	return value
}
