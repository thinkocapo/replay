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

Cronjob
```
# crontab -e
1-59 * * * * cd /home/wcap/thinkocapo/event-maker/ && ./event-to-sentry
# crontab -l
```
## Notes
Mix of strings and serialized data types used in sqlite schema.

The "modified" DSN you initialize sentry_sdk with in event.py will determine which endpoint gets hit in `flask/proxy.py`

`python test.py` and `go run test.go` or for showing the most recent event saved in the database, and total row count.

## TODO

PI  
- event-to-sentry.go parameterize the DATABASE_PATH
- event.go in go
- event.go DSN as Struct with stringify method?
- event-to-sentry.go in go
- Tour of Go

PII
- javascript events
- golang script x events y type
- python. import logger for python
- python. can rename proxy endpoints with /save /forward since the number /2 /3 is really for project Id? confirm it does/nt work

PIII  
- rename 'flask' directory as proxy?
- "save_event" "load_event" or "make pysentry" "make gosentry" and/or "python sentry.py" "go run sentry.go"
- new visual, show Go icon w/ "event-to-sentry" so it's clear the relation between Go->DB->Sentry. don't need 'API/proxy' in step2.
- "github.com/buger/jsonparser" so it'd be bytes->update instead of bytes->interface->update (i.e. it does the Marshalling for me)
- a check to see if Sentry is running? check port:9000 if it's on-premise
- sqlite3 db column for fingerprint so never end up with duplicates
- for non-static languages, log/check the type/class of significant data objects? annotate data type to variable name
- improve variable names. e.g. `request.data` as `request_data_bytes`
- python3 function/class for checking data types  
- before/after hook on Flask endpoint for logging things...name of endpoint
- https://docs.python.org/3/tutorial/modules.html#packages

## Notes
If you think you messed up your database, delete database.db and re-create the file, run db_prep again to set the schema on it.

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

grpc

comparing len(bodyBytes) before/after serialization

```
# didn't look as nice
MODIFIED_DSN_SAVE = ''.join([KEY,'@',SENTRY,'/3'])
MODIFIED_DSN_SAVE = '{KEY}@{PROXY}/3'.format(KEY=KEY,PROXY=PROXY)
```

Unmarshall
https://gobyexample.com/json

Good Marshall Unmarshall examples
https://www.dotnetperls.com/json-go

UUID google package
https://godoc.org/github.com/google/uuid

https://docs.python.org/3/library/typing.html  
https://medium.com/@ageitgey/learn-how-to-use-static-type-checking-in-python-3-6-in-10-minutes-12c86d72677b  

CONVERT 'data' from go object / json into (encoded) ,utf8,bytes,


sqlalchemy==1.3.15


removed flask/.env which had `SQLITE=` in it

https://en.wikipedia.org/wiki/Marshalling_(computer_science)



Go package tests often provide clues as to ways of doing things


For example, from database/sql/sql_test.go,