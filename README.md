<!-- ![The Undertaker](./img/undertaker-1.png) -->
# The Undertaker

<img src="./img/undertaker-4.jpeg" width="450" height="300">  

## why and what's happening?  
Good for test data automation. Do not have to maintain +10 different app and sdk types in Heroku/GCP sending events all the time. Rather, run a single program `event-to-sentry.go` on a cronjob to send those +10 event types for you. It's free. 

<img src="./img/event-maker-slide-2.001.png" width="450" height="300">  

**STEP1**  
`event.py` creates sdk events

`flask/proxy.py` undertakes (intercepts) events on their way to Sentry and saves them in sqlite

**STEP2**  
`event-to-sentry.go` loads events from sqlite and sends them to Sentry, without using an sdk.

## Setup
tested on: ubuntu 18.04 LTS, go 1.12.9 linux/amd64, sentry-sdk 0.14.2, flask Python 3.6.9

use python3 or else else `getvalue()` in `event-to-sentry.py` returns wrong data type ¯\_(ツ)_/¯

1. `git clone getsentry/onpremise` and `./install.sh`
2. DSN's in `.env`, and select right DSN in `proxy.py`, note DSN_REACT vs DSN_FLASK depends on which you're sending through the proxy
3. `pip3 install -r ./python/requirements.txt`
4. `go get github.com/google/uuid github.com/mattn/go-sqlite3 github.com/joho/godotenv github.com/shopspring/decimal`

## Run
Get your proxy and Sentry instance running first.
```
make proxy

# cd getsentry/onpremise
docker-compose up
```
**STEP1**  
```
python3 event.py
```
**STEP2**  
```
go run event-to-sentry.go
go run event-to-sentry.go --id=<id>
go run event-to-sentry.go --all

python3 event-to-sentry.py
python3 event-to-sentry.py <id>
```
See your event in Sentry at `localhost:9000`

**Cronjobs**  
Cronjob on Macbook that sends events in the background
`crontab -e` to open up your Mac's crontab manager
```
# every minute
1-59 * * * * cd /Users/wcap/thinkocapo/undertaker && ./bin/event-to-sentry --all
1-59 * * * * cd /<path>/<to>/undertaker/ && ./event-to-sentry

# every minute, every day of the week M-F
# * * * * 1-5 cd /Users/wcap/thinkocapo/undertaker && ./bin/event-to-sentry --all
# * * * * 1-5 cd /<path>/<to>/undertaker/ && ./event-to-sentry --all

# every 5 minutes
# */5 * * * 1-5 cd /Users/wcap/thinkocapo/undertaker && ./bin/event-to-sentry-neil --all
# */5 * * * 1-5 cd /<path>/<to>/undertaker/ && ./event-to-sentry --all

# crontab -l, to list cronjobs
```

https://crontab.guru/

## Notes
See `python/event.py` for how to construct the 3 'MODIFIED' DSN types which decide which of the 3 endpoints in `proxy.py` which you can hit. Use any app+sdk with one of these MODIFIED_DSN's following the convention in proxy.py

`python3 test/db.py` and `go run test/db.go` are for showing total row count and most recent event.

The timestamp from `go run event-to-sentry.go` is sometimes earlier than today's date

Borrowed code from: getsentry/sentry-python, getsentry/sentry-go, goreplay

https://develop.sentry.dev/sdk/store for info on what the real Sentry endpoints are doing

https://develop.sentry.dev/sdk/event-payloads/ for what a sdk event looks like. Here's [/img/example-payload.png](./img/example-payload.png) from javascript

6 events in the am-transactions-sqlite.db was 57kb
19 events tracing-example was 92kb

`go build -o bin/event-to-sentry-<name> event-to-sentry.go` for who it's for

to use with getsentry/tracing-example, serve the python/proxy.py via `ngrok http 3001` and put the mapped URL in tracing-example's .env like:  
`SENTRY_DSN=https://1f2d7bf845114ba6a5ba19ee07db6800@5b286dac3e72.ngrok.io/3`

## Todo
- database path needs be read from .env by test/*, Makefile, proxy and event-to-sentry
- test that the check for empty DSN  is foolproof
- tracing-example to 3 different Python DSN's
- break up event-to-sentry.go into different files (Package?) e.g. timestamp functions

- Android errors/crashes/sessions
- set Release according to CalVer
- sqlite model needs 'platform, eventType', instead of 'name, type'
- read sentry_key from X-Sentry-Auth, assuming it's always there, and could check in .env for which DSN to use.
- improve log formats and error handling
