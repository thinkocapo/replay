# Replay
Replay is an event traffic replay service. It was formerly presented as [The Undertaker](https://www.youtube.com/watch?v=4QEYJXjC4Jk) at Hackweek 2020.

## Overview  
`main.go` grabs a bunch of sentry events in JSON form from Cloud Storage, updates them, and sends them to your DSN keys at Sentry.io.

## Setup

1. Create a config.yaml. See config.yaml.example.

    `SENTRY_JOB_MONITOR` is good for capturing errors in Replay if you're running it on a job.  
    `DESTINATIONS` are the dsn keys you want to send events to  
    `SOURCES` are sentry org slugs you can scrape events from, and send to your DESTINATIONS. This is by default turned off. See demo_automation.go

2. 
    Install Google Cloud SDK 303.0.0. This is for reading JSON's from the thousands we have crowdsourced and saved in Cloud Storage.

    Obtain the Google Application Credentials file, and put the file path to it in your config.yaml's `GOOGLE_APPLICATION_CREDENTIALS`.

## Run
1. `go build -o bin/main *.go`
2. `./bin/main`
2. Look for your events in your projects on Sentry.io.

## Notes
The `-i` flag is for ignoring the http call to Sentry in `./bin/main -i`

You can pass a `prefix` for the files you want to read from Cloud Storage

Build and run quickly:
```
go build -o bin/main *.go && ./bin/main
```

[Sentry Developer Documentation](https://develop.sentry.dev/sdk/store)

This project was originally called Undertaker because it featured a proxy middleman that captured events on their way to Sentry. Today, we take the event JSON's from Sentry.io and place them in Cloud Storage.