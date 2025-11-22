// main.go - main file
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import "github.com/bassosimone/weekly/internal/cli"

// accessible by tests to mock
var cliMain = cli.Main

func main() {
	cliMain()
}
