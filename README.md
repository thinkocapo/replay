# event-maker
## Goal
Test data automation.

Sending diverse events from multiple sdk types to a Sentry Organization on a regular basis.

Keep 1 app running instead of 1 app per platform.


## What's Happening
<img src="./img/workflow-diagram.jpeg" width="450" height="300">  

STEP1 - Sentry sdk's send events and the API in /flask serves as a proxy to intercept the events and save copies of them in a database. This is useful because apps w/ sdk's do not have to stay running on a scheduled job to keep creating more errors and events. Events are instead saved in a database.

STEP2 - Events do not have to be created because they're alread stored in a database. Load the events from the database and send them to Sentry. This can run on a scheduled job. Sentry sees them as coming from live apps.

[example payload structure](./img/payload-structure.png) from a sentry sdk event:  

## Versions
tested on ubuntu 18.04 LTS

go version go1.12.9 linux/amd64

sentry-sdk==0.14.2

## Install

install -r requirements.txt

## Database
1.
```
docker run -it --rm \
    --name db-postgres \
    -e POSTGRES_PASSWORD=admin \
    -e POSTGRES_USER=admin \
    -p 5432:5432 \
    postgres
```
sudo lsof -i -P -n

sudo service postgresql stop

^ or could do 8080:5432 and it wouldn't conflict?

2.
`docker exec -it db-postgres psql -U admin`

```
\l list databases
\c choose db
\dt list data tables
```

3.
```
CREATE TABLE events(
   pk SERIAL PRIMARY KEY,
   type varchar(40) NOT NULL,
   name varchar(40) NOT NULL,
   data bytea
);
```

create user admin with login password 'admin';

## Run

#### works
Sentry sdk sends events to a Flask API (like a proxy or interceptor) which then sends them to Sentry On-premise
1. `docker-compose up` your getsentry/onpremise, it defaults to localhost:9000
2. `docker run...` the database
2. `make` runs Flask server
3. `python app.py`
4. `localhost:9000` to see your Sentry onprem event

Workflow:  
`python app.py` sdk sends event to the intercetpor.

The `DSN` that you use in your `app.py` determine what the proxy will do. They are mapped to different endpoints in `flask/app.py`.

`localhost:3001/impersonator` will load an event from the database and forward it (by http) to your Sentry instance.



## Gor Middleware
There is a `middleware.go` in this project that's for for sniffing events traffic on the port that Sentry is listening on. It is not a proxy. It is not fully working yet.

#### Install
If using middleware.go then you need gor (goreplay)

1. download gor executable and put to cwd or add it to your $PATH  
https://github.com/buger/goreplay/releases/tag/v1.0.0
2.
```
go get github.com/buger/goreplay/proto  
go get github.com/buger/jsonparser
```

and

install -r requirements.txt

#### Run
Send events using app.py to your on-prem instance. the middleware.go sniffs the events and doesn't interrupt them like a proxy does.   
1. `docker-compose up`
2. `go build middleware.go`
3. `sudo ./gor --input-raw :9000 --middleware "./middleware" --output-http http://localhost:9000/api/2/store`
3. `python3 app.py` using ORIGINAL_DSN

and

https://github.com/getsentry/sentry-python  
https://github.com/getsentry/sentry-go  
https://github.com/getsentry/onpremise  
Borrowed code from https://github.com/getsentry/gor-middleware/blob/master/auth.go

https://github.com/buger/jsonparser

I used this as my 'middleware.go' and removed what I didn't need:  
https://github.com/buger/goreplay/blob/master/examples/middleware/token_modifier.go

About the middleware technique  
https://github.com/buger/goreplay/tree/master/middleware

## Notes
https://flask.palletsprojects.com/en/1.1.x/api/  
https://requests.readthedocs.io/en/master/  
https://realpython.com/python-requests/#request-headers  

json.loads(r.data.decode('utf-8'))['headers']  
request.headers is a #dict  
request.data keys are exception, server_name, tags, event_id, timestamp, extra, modules, contexts, platform, breadcrumbs, level, sdk  
request.data <class 'bytes'>  
request.content_encoding gzip  
request.content_type application/json  
body.getvalue() is a #str or <class 'bytes'>  

This 'DumpRequest' (deprecated/dump-request.go) would be perfect if I could make sentry_sdk send events to a URL of my choosing. Downside is the events would never reach my on-prem Sentry. Maybe support both techniques in this repo:  
https://rominirani.com/golang-tip-capturing-http-client-requests-incoming-and-outgoing-ef7fcdf87113

https://golang.org/pkg/net/http/#Request  
https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body using encoding/json instead of buger/jsonparser  
gor file-server 8000

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
```

replaying the payload many times. grpc


MemoryView  
https://www.postgresql.org/message-id/25EDB20679154BDBB3CBBD335184E1D7%40frank  
https://www.postgresql.org/message-id/C2C12FD0FCE64CE8BB77765A526D3C73%40frank  


"Q. How to save a instance of a Class to the DB?"
"A. You can't store the object itself in the DB. What you do is to store the data from the object and reconstruct it later."
https://stackoverflow.com/questions/2047814/is-it-possible-to-store-python-class-objects-in-sqlite


Troubleshoot - compare len(bytes) on the way in as when it came out...