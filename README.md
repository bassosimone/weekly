# Weekly

![Mascot](docs/weekly.png)

Track your weekly activities using Google Calendar.

See [docs/tutorial.md](docs/tutorial.md) for details about how
to format your calendar entries for tracking.

## Setup

1. You need Go >= 1.24

2. Install this tool

```bash
go install github.com/bassosimone/weekly@latest
```

3. Create a project in Google Cloud Engine (e.g., `weekly`)

4. Create a service account with no permissions within the project

5. Create a JSON key for the service account

6. Create the working directory

```bash
install -d $HOME/.config/weekly
```

7. Move the service account JSON key file to `$HOME/.config/weekly/credentials.json`

8. Share the calendar with the service account email address

9. Take note of the calendar ID

10. Save the calendar ID into the weekly tool by running `weekly init`:

```bash
$HOME/go/bin/weekly init
```

## Usage

Use the `weekly ls` command to list calendar events:

```bash
~/go/bin/weekly ls
```

Use `weekly tutorial` to understand how to format calendar
events or read [docs/tutorial.md](docs/tutorial.md):

```bash
~/go/bin/weekly tutorial
```

Use `weekly --help` to get interactive help:

```bash
~/go/bin/weekly --help
```

## Directories and Files

The `weekly` tool honors `$XDG_CONFIG_HOME`. When `$XDG_CONFIG_HOME`
is not set, we use `$HOME/.config` as the default value.

We save data in `$XDG_CONFIG_HOME/weekly`:

1. `calendar.json` contains the calendar ID

2. `credentials.json` contains the service-account credentials

Every command accepts the `--config-dir <dir>` flag to override the
directory containing the configuration.
