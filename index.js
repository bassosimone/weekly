// This software is free software. See AUTHORS and LICENSE for more
// information on the copying conditions.

"use strict";

const fs = require("fs");
const program = require("commander");
const querystring = require("querystring");
const https = require("https");
const moment = require("moment");
const readline = require("readline");

/*
   _
  (_)___  ___  _ __
  | / __|/ _ \| '_ \
  | \__ \ (_) | | | |
 _/ |___/\___/|_| |_|
|__/
*/

// Transform try..catch json code into a monad
function json_monad(data, callback) {
    try {
        callback(null, JSON.parse(data));
    } catch (error) {
        callback(error);
    }
}

// Read json file, parse the content, and pass object to callback
function json_read_file(path, callback) {
    fs.readFile(path, function(error, data) {
        if (error) {
            callback(error);
            return;
        }
        json_monad(data, callback);
    });
}

// Write json data to the specified file
function json_write_file(path, data, callback) {
    fs.writeFile(path, JSON.stringify(data, undefined, 4) + "\n", callback);
}

// Make an https request expecting a json response
function json_request(options, callback, request_body) {
    let request = https.request(options, function(response) {
        if (response.statusCode !== 200) {
            response.resume();
            if (response.statusCode === 401) {
                callback(new Error("json-request-unauthorized"));
            }
            callback(new Error("json-request-failed"));
            return;
        }
        let response_body = "";
        response.on("data", function(data) { response_body += data; });
        response.on("end", function() { json_monad(response_body, callback); });
    });
    request.on("error", function(error) { callback(error); });
    if (request_body) {
        request.end(request_body);
    } else {
        request.end();
    }
}

/*
                   _   _     ____
  ___   __ _ _   _| |_| |__ |___ \
 / _ \ / _` | | | | __| '_ \  __) |
| (_) | (_| | |_| | |_| | | |/ __/
 \___/ \__,_|\__,_|\__|_| |_|_____|
*/

// Obtain device authentication code from client-id and scope
function oauth2_obtain_user_code(app_path, callback) {
    json_read_file(app_path, function(error, auth) {
        if (error) {
            callback(error);
            return;
        }
        json_request(
            {
              hostname : "accounts.google.com",
              port : 443,
              method : "POST",
              path : "/o/oauth2/device/code",
              headers : {
                  "Content-Type" : "application/x-www-form-urlencoded",
              },
            },
            callback, querystring.stringify({
                "client_id" : auth.client_id,
                "scope" : "https://www.googleapis.com/auth/calendar.readonly",
            }));
    });
}

// Obtain tokens from client-id, client-secret, and device-code
function oauth2_obtain_tokens(app_path, device_path, callback) {
    json_read_file(app_path, function(error, app_info) {
        if (error) {
            callback(error);
            return;
        }
        json_read_file(device_path, function(error, device_info) {
            if (error) {
                callback(error);
                return;
            }
            json_request(
                {
                  hostname : "www.googleapis.com",
                  port : 443,
                  method : "POST",
                  path : "/oauth2/v4/token",
                  headers : {
                      "Content-Type" : "application/x-www-form-urlencoded",
                  },
                },
                callback, querystring.stringify({
                    client_id : app_info.client_id,
                    client_secret : app_info.client_secret,
                    code : device_info.device_code,
                    grant_type : "http://oauth.net/grant_type/device/1.0",
                }));
        });
    });
}

// Refresh cached oauth2 token
function oauth2_refresh(app_path, tokens_path, callback) {
    json_read_file(app_path, function(error, app_info) {
        if (error) {
            callback(error);
            return;
        }
        json_read_file(tokens_path, function(error, tokens_info) {
            if (error) {
                callback(error);
                return;
            }
            json_request(
                {
                  hostname : "www.googleapis.com",
                  port : 443,
                  method : "POST",
                  path : "/oauth2/v4/token",
                  headers : {
                      "Content-Type" : "application/x-www-form-urlencoded",
                  },
                },
                callback, querystring.stringify({
                    client_id : app_info.client_id,
                    client_secret : app_info.client_secret,
                    refresh_token : tokens_info.refresh_token,
                    grant_type : "refresh_token",
                }));
        });
    });
}

/*
           _                _
  ___ __ _| | ___ _ __   __| | __ _ _ __
 / __/ _` | |/ _ \ '_ \ / _` |/ _` | '__|
| (_| (_| | |  __/ | | | (_| | (_| | |
 \___\__,_|_|\___|_| |_|\__,_|\__,_|_|
*/

// Lists the available calendars
function calendar_list(tokens_path, callback) {
    json_read_file(tokens_path, function(error, tokens_info) {
        if (error) {
            callback(error);
            return;
        }
        const options = {
            hostname : "www.googleapis.com",
            port : 443,
            method : "GET",
            path : "/calendar/v3/users/me/calendarList",
            headers : {
                "Authorization" : "Bearer " + tokens_info.access_token,
            },
        };
        json_request(options, function(error, response) {
            if (error) {
                callback(error);
                return;
            }
            callback(null, response);
        });
    });
}

// Get calendar events
function calendar_events(tokens_path, calendar_path, callback) {
    json_read_file(tokens_path, function(error, tokens_info) {
        if (error) {
            callback(error);
            return;
        }
        json_read_file(calendar_path, function(error, calendar_info) {
            if (error) {
                callback(error);
                return;
            }
            const path =
                "/calendar/v3/calendars/" + calendar_info + "/events" + "?" +
                querystring.stringify({
                    timeMin :
                        moment().locale("it").startOf('week').toISOString(),
                    maxResults : 2500,
                });
            const options = {
                hostname : "www.googleapis.com",
                port : 443,
                method : "GET",
                path : path,
                headers : {
                    "Authorization" : "Bearer " + tokens_info.access_token,
                },
            };
            json_request(options, function(error, response) {
                if (error) {
                    callback(error);
                    return;
                }
                callback(null, response);
            });
        });
    });
}

/*
                   _    _
__      _____  ___| | _| |_   _
\ \ /\ / / _ \/ _ \ |/ / | | | |
 \ V  V /  __/  __/   <| | |_| |
  \_/\_/ \___|\___|_|\_\_|\__, |
                          |___/
*/

// Filter available calendars to only return interesting fields
function weekly_filter_calendars(calendars) {
    let result = [];
    for (let index = 0; index < calendars.items.length; ++index) {
        const current = calendars.items[index];
        result.push({
            summary : current.summary,
            id : current.id,
        });
    }
    return result;
}

// Filter calendar events to only return interesting fields
function weekly_filter_events(events) {
    let result = [];
    for (let index = 0; index < events.items.length; ++index) {
        const current = events.items[index];
        result.push({
            summary : current.summary,
            start : current.start.dateTime,
            end : current.end.dateTime,
        });
    }
    return result;
}

// Aggregate calendar events to produce statistics
function weekly_aggregate_events(events) {
    let res = {
        details: {},
        percentage: {},
        total: 0.0,
    };
    for (let index = 0; index < events.length; ++index) {
        const evt = events[index];
        const diff = moment(evt.end).diff(moment(evt.start), "hours", true);
        res.total += diff;
        res.details[evt.summary] = (res.details[evt.summary] | 0.0) + diff;
    }
    Object.keys(res.details).forEach(function (key) {
        res.percentage[key] = (res.details[key] / res.total) * 100.0;
    });
    return res;
}

/*
                 _
 _ __ ___   __ _(_)_ __
| '_ ` _ \ / _` | | '_ \
| | | | | | (_| | | | | |
|_| |_| |_|\__,_|_|_| |_|
*/

const app_path = "private/app.json";
const calendar_path = "private/calendar.json";
const device_path = "private/device.json";
const doc_url = "https://github.com/bassosimone/weekly#create-privateappjson-using-google-developers-console"
const tokens_path = "private/tokens.json";

// Initiate authentication process by requesting a device code to google
function main_init() {
    oauth2_obtain_user_code(app_path, function(error, response) {
        if (error) {
            if (error.code === 'ENOENT' && error.syscall === 'open') {
                if (error.path === app_path) {
                    console.error("fatal: missing file: '" + error.path + "'");
                    console.log("You should create your own app");
                    console.log("See <" + doc_url + "> for instructions");
                    process.exit(1);
                }
            }
            throw error;
        }
        json_write_file(device_path, response, function(error) {
            if (error) {
                throw error;
            }
            console.log("Written device-info at '" + device_path + "'");
            console.log("Now go to <" + response.verification_url + "> and " +
                        "authenticate using " + response.user_code);
            console.log("Then, run 'node index.js --step2'");
        });
    });
}

// After user authorized the app with browser, call this function to get
// a real authentication token to effectively access calendar api
function main_step2() {
    oauth2_obtain_tokens(app_path, device_path, function(error, response) {
        if (error) {
            if (error.code === 'ENOENT' && error.syscall === 'open') {
                console.error("fatal: missing file: '" + error.path + "'");
                console.log("did you run 'node index.js --init'?");
                process.exit(1);
            }
            throw error;
        }
        json_write_file(tokens_path, response, function(error) {
            if (error) {
                throw error;
            }
            console.log("Written tokens-info at '" + tokens_path + "'");
            console.log("Now, run 'node index.js --step3'");
        });
    });
}

// Allows use to select which calendar to use
function main_step3() {
    calendar_list(tokens_path, function(error, response) {
        if (error) {
            throw error;
        }
        const calendars = weekly_filter_calendars(response);
        let rl = readline.createInterface(process.stdin, process.stdout);
        rl.setPrompt(function() {
            let result = "\nAvailable calendars:\n";
            for (let index = 0; index < calendars.length; ++index) {
                result += "  - id: " + calendars[index].id + "\n";
                result += "    summary: '" + calendars[index].summary + "'\n";
            }
            result += "Which id do you want to use? ";
            return result;
        }());
        rl.prompt();
        rl.on("line", function(line) {
              line = line.trim();
              for (let index = 0; index < calendars.length; ++index) {
                  let cal = calendars[index];
                  if (line === cal.id) {
                      rl = null;
                      json_write_file(calendar_path, cal.id, function(error) {
                          if (error) {
                              throw error;
                          }
                          console.log("Written calendar-info at '" +
                                      calendar_path + "'");
                          console.log("You may now use this app");
                          process.exit(0);
                      });
                      return;
                  }
              }
              console.log("\nError: calendar-id not found: " + line);
              rl.prompt();
          }).on("close", function() { process.exit(0); });
    });
}

// Refresh authentication token after it expired
function main_refresh() {
    oauth2_refresh(app_path, tokens_path, function (error, response) {
        if (error) {
            throw error;
        }
        json_read_file(tokens_path, function (error, tokens_info) {
            if (error) {
                throw error;
            }
            // Replace expired token with new token:
            tokens_info.access_token = response.access_token;
            json_write_file(tokens_path, tokens_info, function (error) {
                if (error) {
                    throw error;
                }
            });
        });
    });
}

// Query the calendar and print statistics
function main_weekly() {
    calendar_events(tokens_path, calendar_path, function(error, response) {
        if (error) {
            if (error.code === 'ENOENT' && error.syscall === 'open') {
                console.error("fatal: missing file: '" + error.path + "'");
                console.log("did you run 'node index.js --init'?");
                process.exit(1);
            }
            if (error.message === 'json-request-unauthorized') {
                console.error("fatal: you are not authorized");
                console.log("Try running 'node index.js --refresh'");
                process.exit(1);
            }
            throw error;
        }
        console.log(weekly_aggregate_events(weekly_filter_events(response)));
    });
}

program.version("1.0.0")
    .option("--init", "Triggers the initialization procedure")
    .option("--refresh", "Refresh authentication when not authorized")
    .option("--step2", "Second of initialization procedure")
    .option("--step3", "Third step of initialization procedure")
    .parse(process.argv);

if (program.init) {
    main_init();
} else if (program.step2) {
    main_step2();
} else if (program.step3) {
    main_step3();
} else if (program.refresh) {
    main_refresh();
} else {
    main_weekly();
}
