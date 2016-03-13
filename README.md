# Weekly

Track your weekly activities using Google Calendar.

Uses ["OAuth2 for devices" authentication
flow](https://developers.google.com/identity/protocols/OAuth2ForDevices).

Requires Node 5.5+ (uses some ECMA 6 features).

Instructions to get you going:

## Clone repository, install dependencies

As a first step clone the repository and install dependencies

```
git clone https://github.com/bassosimone/weekly
cd weekly
npm install
```

## Create private/app.json using Google Developers Console

Now you need to create your own application using Google Developers Console.

Follow these steps:

- Go to https://console.developers.google.com

- Create new project called `weekly`

- Enable the `Calendar API` for such project

- Create `OAuth Client ID` credentials

- Select type `Other`

- Create file `private/app.json` according to this template

```json
{
  "client_id": "put-client-id-here",
  "client_secret": "put-client-secret-here"
}
```

## Authenticate device for using the Calendar API

Now we need to register this application for using the Calendar API. This
is a multi-step process, started with this command:

```
node index.js --init
```

You will need to go to a Google website indicated by the program and enter
an authentication code indicated by the program.

## Complete authentication for using Calendar API

Once you've inserted the code into the indicated Google website, run

```
node index.js --step2
```

After this step, you are authenticated and can call the Calendar API.

## Select the calendar you want to use

The following command allows you to choose the calendar you want to use:

```
node index.js --step3
```

Specifically, the program shows you the list of available calendars and then
you shall tell it which calendar-id you want to use.

## Query your calendar

To query your calendar, use this command:

```
node index.js
```

By default it returns statistics related to the last week.

## Refresh your token

The token obtained using `--step2` should typically expire after an hour.

If you try to query the calendar when the token is expired, Google API will
reply with 401 and the program will tell you so and suggest to run:

```
node index.js --refresh
```

This *should* refresh the token. There is a maximum number of times you
can refresh a token. Afterwards, my understanding is that you shall restart
the initialization procedure from `--init`.

