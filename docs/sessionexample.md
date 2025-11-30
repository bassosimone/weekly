# Weekly Session Example

The day is over. I run `weekly ls` to see what I have done today:

```
$ weekly ls
┌──────────────────┬───────┬──────────┬─────────────┬─────────┬─────────┐
│    START TIME    │ HOURS │ PROJECT  │  ACTIVITY   │  TAGS   │ PERSONS │
├──────────────────┼───────┼──────────┼─────────────┼─────────┼─────────┤
│ 2025-11-22 10:00 │ 0.5   │ nexa     │ meeting     │ staff   │ all     │
│ 2025-11-22 10:45 │ 0.8   │ nexa     │ development │ neubot  │         │
│ 2025-11-22 11:45 │ 1.0   │ nexa     │ development │ neubot  │         │
│ 2025-11-22 14:30 │ 0.5   │ mlab     │ meeting     │ standup │ all     │
│ 2025-11-22 15:15 │ 1.2   │ mlab     │ development │ locate  │         │
│ 2025-11-22 16:45 │ 0.8   │ mlab     │ development │ iqb     │         │
│ 2025-11-22 19:15 │ 1.0   │ personal │ development │ weekly  │         │
└──────────────────┴───────┴──────────┴─────────────┴─────────┴─────────┘
```

I want to know how many hours I have spent per `$project`:

```
$ weekly ls --aggregate daily
┌──────────────────┬───────┬──────────┬──────────┬──────┬─────────┐
│    START TIME    │ HOURS │ PROJECT  │ ACTIVITY │ TAGS │ PERSONS │
├──────────────────┼───────┼──────────┼──────────┼──────┼─────────┤
│ 2025-11-22 00:00 │ 2.5   │ mlab     │          │      │         │
│ 2025-11-22 00:00 │ 2.2   │ nexa     │          │      │         │
│ 2025-11-22 00:00 │ 1.0   │ personal │          │      │         │
└──────────────────┴───────┴──────────┴──────────┴──────┴─────────┘
```

I can obtain the same information as a CSV:

```
$ weekly ls --aggregate daily --format csv
2025-11-22T00:00:00Z,2h30m0s,mlab,,,
2025-11-22T00:00:00Z,2h15m0s,nexa,,,
2025-11-22T00:00:00Z,1h0m0s,personal,,,
```

Or as JSON:

```
% weekly ls --aggregate daily --format json|jq
{
  "Project": "mlab",
  "Activity": "",
  "Tags": null,
  "Persons": null,
  "StartTime": "2025-11-22T00:00:00Z",
  "Duration": 9000000000000
}
{
  "Project": "nexa",
  "Activity": "",
  "Tags": null,
  "Persons": null,
  "StartTime": "2025-11-22T00:00:00Z",
  "Duration": 8100000000000
}
{
  "Project": "personal",
  "Activity": "",
  "Tags": null,
  "Persons": null,
  "StartTime": "2025-11-22T00:00:00Z",
  "Duration": 3600000000000
}
```

Since I am at the end of the week, I want now to see what I have
actually done overall this week (i.e., in the past 5 days)

```
$ weekly ls --aggregate daily --days 5
┌──────────────────┬───────┬──────────┬──────────┬──────┬─────────┐
│    START TIME    │ HOURS │ PROJECT  │ ACTIVITY │ TAGS │ PERSONS │
├──────────────────┼───────┼──────────┼──────────┼──────┼─────────┤
│ 2025-11-18 00:00 │ 0.5   │ nexa     │          │      │         │
│ 2025-11-19 00:00 │ 3.0   │ nexa     │          │      │         │
│ 2025-11-19 00:00 │ 5.0   │ mlab     │          │      │         │
│ 2025-11-20 00:00 │ 8.5   │ mlab     │          │      │         │
│ 2025-11-21 00:00 │ 2.5   │ nexa     │          │      │         │
│ 2025-11-22 00:00 │ 2.5   │ mlab     │          │      │         │
│ 2025-11-22 00:00 │ 2.2   │ nexa     │          │      │         │
│ 2025-11-22 00:00 │ 1.0   │ personal │          │      │         │
└──────────────────┴───────┴──────────┴──────────┴──────┴─────────┘
```

To get totals, I can also aggregate weekly:

```
$ weekly ls --aggregate weekly --days 5
┌──────────────────┬────────┬──────────┬──────────┬──────┬─────────┐
│    START TIME    │ HOURS  │ PROJECT  │ ACTIVITY │ TAGS │ PERSONS │
├──────────────────┼────────┼──────────┼──────────┼──────┼─────────┤
│ 2025-11-17 00:00 │ 8.2    │ nexa     │          │      │         │
│ 2025-11-17 00:00 │ 16.0   │ mlab     │          │      │         │
│ 2025-11-22 00:00 │ 1.0    │ personal │          │      │         │
└──────────────────┴────────┴──────────┴──────────┴──────┴─────────┘
```

I can also filter by project

```
$ weekly ls --aggregate weekly --days 5 --project nexa
┌──────────────────┬────────┬──────────┬──────────┬──────┬─────────┐
│    START TIME    │ HOURS  │ PROJECT  │ ACTIVITY │ TAGS │ PERSONS │
├──────────────────┼────────┼──────────┼──────────┼──────┼─────────┤
│ 2025-11-17 00:00 │ 8.2    │ nexa     │          │      │         │
└──────────────────┴────────┴──────────┴──────────┴──────┴─────────┘
```

I can also filter by tab

```
$ weekly ls --tag neubot
┌──────────────────┬───────┬──────────┬─────────────┬─────────┬─────────┐
│    START TIME    │ HOURS │ PROJECT  │  ACTIVITY   │  TAGS   │ PERSONS │
├──────────────────┼───────┼──────────┼─────────────┼─────────┼─────────┤
│ 2025-11-22 10:45 │ 0.8   │ nexa     │ development │ neubot  │         │
│ 2025-11-22 11:45 │ 1.0   │ nexa     │ development │ neubot  │         │
└──────────────────┴───────┴──────────┴─────────────┴─────────┴─────────┘
```
