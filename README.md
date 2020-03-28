# event-maker

## What's Happening

Sentry_sdk sends events and the API in /flask 'intercepts' them like a proxy, then forwards them along to a Sentry On-Prem instance

The middleware.go is for sniffing traffic on the Sentr On-Prem's listening port, but isn't fully working yet.

example payload structure from a sentry sdk event:  
![payload-structure](./img/payload-structure.png)

## Versions
tested on ubuntu 18.04 LTS

go version go1.12.9 linux/amd64

sentry-sdk==0.14.2

## Install
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
2. `make` runs Flask server
3. `python app.py` using MODIFIED_dsn
4. `localhost:9000` to see your Sentry onprem event

#### doesnt' work yet
Send events using app.py to your on-prem instance. the middleware.go sniffs the events and doesn't interrupt them like a proxy does.   
1. `docker-compose up`
2. `go build middleware.go`
3. `sudo ./gor --input-raw :9000 --middleware "./middleware" --output-http http://localhost:9000/api/2/store`
3. `python3 app.py` using ORIGINAL_DSN


## Notes
#### Sentry & buger's goreplay
https://github.com/getsentry/sentry-python  
https://github.com/getsentry/sentry-go  
https://github.com/getsentry/onpremise  
Borrowed code from https://github.com/getsentry/gor-middleware/blob/master/auth.go

https://github.com/buger/jsonparser

I used this as my 'middleware.go' and removed what I didn't need:  
https://github.com/buger/goreplay/blob/master/examples/middleware/token_modifier.go

About the middleware technique  
https://github.com/buger/goreplay/tree/master/middleware

#### other
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