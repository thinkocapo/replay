# Replay
Replay is an event traffic replay service. It was formerly presented as [The Undertaker](https://www.youtube.com/watch?v=4QEYJXjC4Jk) at Hackweek 2020.

## Why?  
`main.go` takes a bunch of JSON events, modifies them, and sends them up to Sentry. No SDK involved.

## Setup

Install Google Cloud SDK 303.0.0. This is for reading JSON's from the thousands we have saved in Cloud Storage.

https://develop.sentry.dev/sdk/store

1. Create a config.yaml. See config.yaml.example.

    SENTRY_JOB_MONITOR is good for capturing errors in Replay if you're running it on a job.  
    DESTINATIONS are the dsn keys you want to send events to  
    SOURCES are sentry org slugs you can scrape events from, and send to your DESTINATION dsn keys  

## Run
1. `go build -o bin/main *.go && ./bin/main`
2. `./bin/main`
2. Look for your events in your projects on Sentry.io.

## Notes
`-i` is for ignoring the http call to Sentry in `./bin/main -i`

you can pass a `prefix` for the files you want to read from Cloud Storage
```
go build -o bin/main *.go && ./bin/main <prefix> -i
```


