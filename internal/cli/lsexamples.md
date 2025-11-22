## Examples

To see what you have done today in a user friendly format use:

    weekly ls

To get the same data in a format suitable for invoicing:

    weekly ls --format invoice --aggregate daily

You can also change the format to be JSON:

    weekly ls --format json

Alternatively, you can change the format to be CSV:

    weekly ls --format csv

You can go back in time with the `--days` flag:

    weekly ls --days 3

You can aggregate daily and by project with `--aggregate`:

    weekly ls --days 3 --aggregate daily

You can also aggregate monthly:

    weekly ls --days 60 --aggregate monthly

You can also compute the total in the aggregation period:

    weekly ls --total

The `invoice` format is a simplified CSV format suitable
for generating invoices.
