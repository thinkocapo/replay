# The Undertaker
The Undertaker is an event traffic replay service.

## Why?  
`main.go` takes a bunch of JSON events, modifies them, and sends them up to Sentry. No SDK involved.

## Setup

1. Enter your DSN's in `.env`  
```
DSN_JAVASCRIPT_SAAS=
DSN_PYTHON_SAAS=
```

## Run Locally
1. `make` or `go build -o bin/main *.go && ./bin/main`
2. Look for your events in Sentry

## Notes
-i is for ignoring the http call to Sentry
```
go build -o bin/main *.go && ./bin/main -i
```
install Google Cloud SDK 303.0.0  
https://develop.sentry.dev/sdk/store
