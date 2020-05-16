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

# Go works more consistently
go run event-to-sentry.go
go run event-to-sentry.go --all
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
- event.go in go

- event.go DSN as Struct with stringify method?
- event-to-sentry.go var DATABASE_PATH
- event-to-sentry.go DSN method for SENTRY_URL

- proxy.py in .go



http://127.0.0.1:3999/basics/5 function (x,y int) params if both are ints
http://127.0.0.1:3999/basics/6 func return multiple values
http://127.0.0.1:3999/basics/7 'empty returns' if you say func returns "named return values" like (x,y int) <--enforces return values are specific variables treated within the function. this is like a good function signature. Overloaded functions

var c bool <-- defaults as false
var i int <-- defaults as 0

var c, python, java = true, false, "no!"

Type Conversions The expression T(v) converts the value v to the type T. so try int("5")

var i int
j := i // j is an int