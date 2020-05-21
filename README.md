<!-- ![The Undertaker](./img/undertaker-1.png) -->
# The Undertaker

<img src="./img/undertaker-4.jpeg" width="450" height="300">  

## why?  
Good for test data automation. Do not have to maintain +10 different app and sdk types in Heroku/GCP sending events all the time. Rather, run a single program `event-to-sentry.go` on a cronjob to send those +10 event types for you. It's free. 

## what's happening
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
2. DSN in `.env`
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
Cronjob on your Macbook that sends events in the background
```
# crontab -e
1-59 * * * * cd /home/wcap/thinkocapo/event-maker/ && ./event-to-sentry
# crontab -l
```

## Notes
There are 3 modified DSN's in `event.py` that correspond to the 3 different endpoints in `flask/proxy.py` which you can hit.`

`python test/db.py` and `go run test/db.go` are for showing total row count and most recent event.

The timestamp from `go run event-to-sentry.go` is sometimes earlier than today's date

This repo borrowed code from: getsentry/sentry-python's transport.py, core_api.py, event_manager.py, and getsentry/sentry-go

[/img/example-payload.png](./img/example-payload.png) from a sentry sdk event

## Todo
- tour of go - methods&interfaces

- send go events to proxy.py? (yes do first, test can test scripting of python+go events together)
    - script for sending python+go events together to Sentry:9000
'OR'
- send python events to proxy.go? (NEED create proxy.go) (yes and handle different compressions here, rather than do that in proxy.py)


- event-to-sentry.go var DATABASE_PATH

- have all developers use a DSN that points to a cloud hosted proxy)
