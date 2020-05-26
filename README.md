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
3. `pip3 install -r ./flask/requirements.txt`
4. `go get github.com/google/uuid github.com/mattn/go-sqlite3 github.com/joho/godotenv`

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
go run event-to-sentry.go --all

python3 event-to-sentry.py
python3 event-to-sentry.py <id>
```
See your event in Sentry at `localhost:9000`

**OPTIONAL**  
Cronjob on your Macbook that sends events in the background. Still needs sentry-cli usage for setting the release, and then event-to-sentry.go to use that same release.
```
# crontab -e
1-59 * * * * cd /home/wcap/thinkocapo/event-maker/ && ./event-to-sentry
# crontab -l
```

## Notes
There are 3 modified DSN's in `event.py` that correspond to the 3 different endpoints in `flask/proxy.py` which you can hit.`

`python test/db.py` and `go run test/db.go` are for showing total row count and most recent event.

The timestamp from `go run event-to-sentry.go` is sometimes earlier than today's date

https://develop.sentry.dev/sdk/store for info on what the real Sentry endpoints are doing

This repo borrowed code from: getsentry/sentry-python's transport.py, core_api.py, event_manager.py, and getsentry/sentry-go

[/img/example-payload.png](./img/example-payload.png) from a sentry sdk event

## Todo

- sentry-cli for Release for js events from Database, so they're minified
- sentry-cli should create a release and associate commits, use a Release# that relates to day of the week or day/month/year
- when loading events from database, should be able to set this same day/month/year as the release, so it'll get associated in Sentry.io

- Android errors/crashes/sessions

- write a proxy.go, but make sure Mobile stuff works in proxy.py first
- event-to-sentry.go var DATABASE_PATH
- improve use of log.Fatal vs panic, error handling
- add and test 'X-Sentry-Auth' or whatever will get used for ApplicationManagement tracing 
- which request.header indicates what kind of sdk/event it's from? user-agent for now. Or...  
- how to read sentry_key from incoming request at proxy level? so then proxy can check a .env and figure out which DSN (projectId) to send to....
- proxy that any developer can run locally, which forwards to their Sentry of choice, as well as a cloud db ("crowdsourced")
