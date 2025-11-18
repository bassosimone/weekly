// must.go - must0 and must1 func
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import "log"

func must0(err error) {
	if err != nil {
		log.Fatalf("fatal: %s", err.Error())
	}
}

func must1[T any](value T, err error) T {
	must0(err)
	return value
}
