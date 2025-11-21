// must.go - must0 and must1 func
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import "fmt"

func must0(env *execEnv, err error) {
	if err != nil {
		fmt.Fprintf(env.Stderr(), "fatal: %s", err.Error())
		env.Exit(1)
	}
}

func must1[T any](value T, err error) T {
	// Note: here we're using the global execEnv. A bummer but we
	// cannot avoid it without breaking the usage pattern.
	//
	// A possible better solution is to make it possible to support this
	// pattern via a specific `panic` in a new release of clip.
	must0(env, err)
	return value
}
