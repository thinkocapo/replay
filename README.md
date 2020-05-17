# Undertaker
(image)

## What's Happening
<img src="./img/workflow-diagram.jpeg" width="450" height="300">  

Intercept or "undertake" events on their way to Sentry using the proxy API (Flask) and store them in Sqlite. 1 app sends loads them from database and sends to SEntry. Good for test data automation (cronjob below). Apps w/ sdk's do not have to stay running on a scheduled job to keep creating more errors and events

[example payload structure](./img/payload-structure.png) from a sentry sdk event:  

## Versions
tested on ubuntu 18.04 LTS

go version go1.12.9 linux/amd64

sentry-sdk==0.14.2

```
$ flask --version
Python 3.6.9
```

use Python3 for event-to-sentry.py or else BytesIo.getvalue() will return string instead of bytes

## Setup

1. get a DSN from Sentry on localhost:9000 and put it in `.env`

2. `pip3 install -r ./flask/requirements.txt`
```
virtualenv -p /usr/bin/python3 .virtualenv
source .virtualenv/bin/activate
```
3. `git clone getsentry/onpremise` and `install.sh`

4. 
```
go get github.com/google/uuid
go get github.com/mattn/go-sqlite3
go get github.com/joho/godotenv
```
## Run
**STEP1**  
run the proxy (Flask) then send an event to it. proxy's default behavior is to save the event in Sqlite
```
make proxy

python event.py
```
**STEP2**  
Get Sentry running, then load event(s) from Sqlite and send to Sentry
```
# getsentry/onpremise
docker-compose up

# Go works more consistently. takes the last saved event unless you specify --all
go run event-to-sentry.go
go run event-to-sentry.go --all

# takes last saved event unless you specify a <id>
python event-to-sentry.py
python event-to-sentry.py <id>
```
See your event in Sentry at `localhost:9000`

Cronjob (optional)
```
# crontab -e
1-59 * * * * cd /home/wcap/thinkocapo/event-maker/ && ./event-to-sentry
# crontab -l
```

## Notes
The timestamp from `go run event-to-sentry.go` is sometimes earlier than today's date

The "modified" DSN you initialize sentry_sdk with in event.py will determine which endpoint gets hit in `flask/proxy.py`

`python test.py` and `go run test.go` or for showing the most recent event saved in the database, and total row count.

sqlite-prep.py

```
# reference
getsentry/sentry-python transport.py, core_api.py, event_manager.py
getsentry/sentry-go

# request headers that work
{
    'Host': 'localhost:3001',
    'Accept-Encoding': 'identity', 
    'Content-Length': '1501', 
    'Content-Encoding': 'gzip', 
    'Content-Type': 'application/json', 
    'User-Agent': 'sentry.python/0.14.2'
}
```

## Todo
- tour

- send go events to proxy.py? (yes do first, test can test scripting of python+go events together)
    - script for sending python+go events together to Sentry:9000
'OR'
- send python events to proxy.go? (NEED create proxy.go) (yes and handle different compressions here, rather than do that in proxy.py)


- event-to-sentry.go var DATABASE_PATH

