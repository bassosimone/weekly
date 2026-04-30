// main.go - main file
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"

	"github.com/bassosimone/deferexit"
	"github.com/bassosimone/weekly/internal/cli"
)

// accessible by tests to mock
var cliMain = cli.Main

func main() {
	defer deferexit.Recover(os.Exit)
	cliMain()
}
