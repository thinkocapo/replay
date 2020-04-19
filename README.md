# Undertaker
## Goal
Test data automation. Have 1 app send events for all sdk's rather than 1 app per sdk. Prepare these events ahead of time in a database, by intercepting or "undertake them" on their way to Sentry using the proxy API (Flask) in this repo and storing them to sqlite.


## What's Happening
<img src="./img/workflow-diagram.jpeg" width="450" height="300">  

STEP1 - Sentry sdk's send events to the API defined in /flask/server-sqlite.py. It acts like a proxy that intercept the events before they hit Sentry. It saves copies of them in a database. This is useful because apps w/ sdk's do not have to stay running on a scheduled job to keep creating more errors and events. Events are instead saved in a database for replaying in the future.

STEP2 - Events do not have to be created because they're alread stored in a database. Load the events from the database and send them to Sentry. This can run on a scheduled job. Sentry thinks they're coming from live apps.

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

1. put your DSN in `.env`

2. `pip install -r ./flask/requirements.txt`

3. `git clone getsentry/onpremise` and `install.sh`

## Run
Get proxy running (and Sentry running/listening), Send some events to Database via the proxy:
```
# Flask
make proxy


# creates an event, hits an endpoint in Flask, saves event to database
python app.py
```

Get Sentry running, Load events from DB and send to Sentry
```
# getsentry/onpremise
docker-compose up

# script gets event from database and sends to Sentry
python event-to-sentry.py
go run event-to-sentry.go

# See your event in Sentry at `localhost:9000`
```

Note - The modified `DSN` variant that you use when initializing Sentry will determine what the proxy will do. They are mapped to different endpoints in `flask/server-sqlite.py`

Note - `python sqlite-test.py` and `go run sqlite-test.go` show the most recent event from the database

## TODO

PI  
- Tour of Go
- Go - send event to Sentry Instance

PII
- golang scripts. x events y type. release as Day.
- golang script for grabbing x events of type y from DB and send to Sentry,io
- gloang script on a crontab (macbook cronjob) every hour

PIII  
- improve many variable names. e.g. `request.data` as `request_data_bytes`
- Flask response object handling, show status of response and ...'created in Sentry'
- Javascript events
- raise Exception('big problem')
- python3 function/class for checking data types  
https://docs.python.org/3/library/typing.html  
https://medium.com/@ageitgey/learn-how-to-use-static-type-checking-in-python-3-6-in-10-minutes-12c86d72677b  
- before/after hook on Flask endpoint for logging name of endpoint
- better package https://docs.python.org/3/tutorial/modules.html#packages
- new visual
- db column for fingerprint so never end up with duplicates

## Notes
https://flask.palletsprojects.com/en/1.1.x/api/  
https://requests.readthedocs.io/en/master/  
https://realpython.com/python-requests/#request-headers  

request.headers is a #dict  
request.data keys are exception, server_name, tags, event_id, timestamp, extra, modules, contexts, platform, breadcrumbs, level, sdk  

```
getsentry/sentry-python
transport.py, core_api.py, event_manager.py
```

Working Request Headers
```
{
    'Host': 'localhost:3001',
    'Accept-Encoding': 'identity', 
    'Content-Length': '1501', 
    'Content-Encoding': 'gzip', 
    'Content-Type': 'application/json', 
    'User-Agent': 'sentry.python/0.14.2'
}
```

```
type(request) <class 'werkzeug.local.LocalProxy'>
type(request.headers) <class 'werkzeug.datastructures.EnvironHeaders'>
type(request.data) <class 'bytes'>
200 RESPONSE and event_id b'{"id":"2e8e7ab795ed4f9fb70d172aa2b79815"}'

print('request.headers', request.headers) (K | V line separated)
print('type(request.data)', type(request.data)) # <class 'bytes'>
```

replaying the payload many times. grpc



Troubleshoot - compare len(bytes) on the way in as when it came out...

If you think you messed up your database, delete database.db and re-create the file, run db_prep again to set the schema on it.

```
MODIFIED_DSN_SAVE = ''.join([KEY,'@',SENTRY,'/3'])
MODIFIED_DSN_SAVE = '{KEY}@{PROXY}/3'.format(KEY=KEY,PROXY=PROXY)
```

https://gobyexample.com/json