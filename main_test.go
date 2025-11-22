// main_test.go - main tests
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
	// save and restore the globals
	saved := cliMain
	defer func() {
		cliMain = saved
	}()

	// make sure we know if the function is called
	var called bool
	cliMain = func() {
		called = true
	}

	// run the function we're testing
	main()

	// make sure cliMain has been called
	assert.True(t, called)
}
