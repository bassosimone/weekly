# Weekly Tutorial

Create events in the selected calendar spanning the duration of
a specific activity and describe the activity as follows:

    $project %activity #tag @person

Where:

1. `$project` identifies the funding/funded project (e.g., `$mlab`)
and must be specified exactly once for each entry

2. `%activity` identifies the activity (e.g., `%development`) and
must be specified exactly once for each entry

3. `#tag` is optional and may appear zero or more time allowing
to specify more details (e.g., `#iqb #python #streamlit`)

4. `@person` is optional and may appear zero or more time allowing
to specify who helped out (e.g., `@bassosimone @sbs`)

The order with which you specify these tags is irrelevant.

We use the above convention to semantically tag events.

Any additional token will be ignored.

Refer to `weekly ls --help` for help about reading the calendar
or read [lsexamples.txt](lsexamples.txt) directly.
