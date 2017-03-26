# Server Seer

Server seer is a small Go application designed to run ANY shell command to parse
a single value from the output. The output is then stored in an SQLite database.

## Vision

The vision of this application is to make a very simple way to log the system status
but it also could be used for more complex commands as checking Apache status,
number of entries in a database, etc. It could be anything a all, as long as it
is a valid shell command that produces a single number.


## Current status

Currently, there is no functionality to send that data. A sender is being
prepared and will work once I have the possibility to work on the server,
listening for those API calls.

~~Also, a shell UI is being worked on, based on [gizak/termui](https://github.com/gizak/termui).~~
Scrapped for now because of lacking functionality.

A simple local WEB server for viewing the metrics is being worked on.

## Using

### Install

To use the application, you will need to build it. To build it you just need
to [install Go](https://golang.org/doc/install). Once you have Go, just run 
````go build cmd/main.go````, set up your configuration file and you're good to go.

Once you have one build, you can use the same file on all machines with the same
architecture.

### Logging

While the application is designed to handle multiple ways to log, currently it only logs
to screen, so feel free to stream it to a file yourself until file logging is
implemented.

## Configuration

### commands.json

This file contains all the possible commands that _can_ be used in the application.

````
[{
    "name": "disk_space",
    "command": "df | grep '<filesystem>' | tr -s ' '"
}, {
    "name": "disk_space_used_percentage",
    "command": "df | grep '<filesystem>' | tr -s ' ' | tr ' ' '\n' | tail -2 | head -1 | tr -d '%'"
}
````

The file contains a _name_, that can be used in the configuration, and a _command_ 
that is called when the command is used. The command can contain a placeholer,
that would be filled in during setup of the handlers.


### configuration.json

This file defines the whole functionality.

````
{
    "commandsFile": "commands.json",
    "databaseFile": "../storage.db",
    "handlers": [{
        "name": "Load 5min",
        "identifier": "load_5min",
        "command": "system_load",
        "frequency": 10,
        "placeholders": [{
            "name": "timeframe",
            "value": "2"
        }]
    }],
    "sendData": true,
    "senderSettings": {
        "url": "http://www.mocky.io/v2/5185415ba171ea3a00704eed",
        "entriesPerCycle":10,
        "cycleFrequency":30
    }
}
````

* _commandsFile_ - define a custom commands file;
* _databaseFile_ - SQLite file location (can be non existant as long as the location is writable);
* _handlers_ - all the handlers that will be ran
    * _name_ - just a visual name for logging inside the application
    * _identifier_ - the identifier of this specific handler, will be send to the API
    * _command_ - command name from the commands file
    * _frequency_ - how often will this handler be called
    * _placeholders_ - all place holders that exist in the command
        * _name_ - name of the placeholders
        * _value_ - value to replace with
* _sendData_ - whether the sender should be used.
* _senderOptions_ - if sender is enabled, settings will be used for the API
    * _url_ - URL for the API
    * _entriesPerCycle_ - how many entries will be checked on each cycle 
    * _cycleFrequency_ - how often entries will be parsed, and attempted to be sent

## Known and tested commands

I will try to keep and update a list of commands that can be used out-of-the-box
for people who do not want to fiddle with bash too much.

### Disk space percentage

Returns the percentage of how much disk space is left.

````df | grep '<filesystem>' | tr -s ' ' | tr ' ' '\n' | tail -2 | head -1 | tr -d '%'````

_filesystem_ - filesystem name

### System load

CPU load for a specific timeframe.

````uptime | tr ' ' '\n' | tr -d ',' | tail -3 | sed -n '<timeframe>p'````

_timeframe_ - "1" for 1min average; "2" for 5min avg; "3" for 15min avg.

### Remaining memory

Remaining memory in bytes.

````free | grep 'Mem' | tr -s ' ' '\n' | tail -n +2 | sed -n '3p'````