![Mascot](docs/weekly.png)

[![Build Status](https://github.com/bassosimone/weekly/actions/workflows/go.yml/badge.svg)](https://github.com/bassosimone/weekly/actions) [![codecov](https://codecov.io/gh/bassosimone/weekly/branch/master/graph/badge.svg)](https://codecov.io/gh/bassosimone/weekly)

Track your weekly activities using [Google Calendar](https://calendar.google.com/).

See [tutorial.md](internal/cli/tutorial.md) for details about how
to format your calendar entries for tracking.

See [sessionexample.md](docs/sessionexample.md) for an example
of a typical session where I am using `weekly` to track my time usage.

## Install

1. You need Go >= 1.24

2. Install this tool

```bash
go install github.com/bassosimone/weekly@latest
```

## First-Time Setup

The following instructions assume that

```bash
if [[ -n $XDG_CONFIG_HOME ]]; then
	export configDir=$XDG_CONFIG_HOME/weekly
else
	export configDir=$HOME/.config/weekly
fi
```

1. Create a project in Google Cloud Engine (e.g., `weekly`)

2. Create a service account with no permissions within the project

3. Create a JSON key for the service account

4. Create the configuration directory

```bash
install -d $configDir
```

5. Move the service account JSON key file to `$configDir/credentials.json`
and then make it as private as possible:

```bash
chmod 600 $configDir/credentials.json
```

6. Share the calendar with the service account email address

7. Take note of the calendar ID

8. Save the calendar ID into the weekly tool by running `weekly init`
and following its instructions to complete the configuration:

```bash
$HOME/go/bin/weekly init
```

## Usage

Use `weekly tutorial` to understand how to format calendar
events or read [tutorial.md](internal/cli/tutorial.md):

```bash
$HOME/go/bin/weekly tutorial
```

Use the `weekly ls` command to list calendar events:

```bash
$HOME/go/bin/weekly ls
```

Use:

```bash
$HOME/go/bin/weekly ls --help
```

or read [lsexamples.md](internal/cli/lsexamples.md) to
see additional `weekly ls` usage examples.

Use `weekly --help` to get interactive help:

```bash
$HOME/go/bin/weekly --help
```

## Environment Variables

The `weekly` tool honors `$XDG_CONFIG_HOME`. If this variable
is set, the config directory is:

```bash
$XDG_CONFIG_HOME/weekly
```

Otherwise, `weekly` uses this config directory:

```bash
$HOME/.config/weekly
```

Every command accepts the `--config-dir <dir>` flag to override the
directory containing the configuration.

## Files

The `weekly` tool requires two files inside its config directory:

1. `credentials.json` containing the service-account credentials and
manually created during the first-time setup process.

2. `calendar.json` containing the ID of the calendar to use and
created in the first-time setup process by `weekly init`.

## Exit Code

The `weekly` tools exits with `0` on success and nonzero on failure.

## Build From Source

You need Go >= 1.24. Run these commands:

```bash
git clone git@github.com/bassosimone/weekly
cd weekly
go build -v .
```

The `./weekly` binary will be created in the current directory.

## License

```
SPDX-License-Identifier: GPL-3.0-or-later
```

## Dependencies

- [github.com/bassosimone/clip](https://pkg.go.dev/github.com/bassosimone/clip)
- [github.com/google/go-cmp](https://pkg.go.dev/github.com/google/go-cmp)
- [github.com/olekukonko/tablewriter](https://pkg.go.dev/github.com/olekukonko/tablewriter)
- [github.com/rogpeppe/go-internal](https://pkg.go.dev/github.com/rogpeppe/go-internal)
- [github.com/stretchr/testify](https://pkg.go.dev/github.com/stretchr/testify)
- [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2)
- [google.golang.org/api](https://pkg.go.dev/google.golang.org/api)
