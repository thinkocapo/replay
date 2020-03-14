# event-maker

## What's Happening
We self-host a Sentry instance on localhost:9000

We produce errors in app.py and Sentry SDK sends them as events to localhost:9000

We have a Go Replay called 'gor' running to sniff these POST requests hitting localhost:9000

The request body (and possibly headers, etc.) are of interest for analysis. Could write them to a DB, analyze them later.

TODO - create thousands of events via app.py or a homegrown cli tool at once, then run ML on them. and/or could compare them to their post-ingestion state (i.e. where they're stored in Sentry.io/snuba). This cli testing tool is something i've been intersted in developing for a while, for populating test data, aside from ML.  

THOUGHT - I realized mid-way through this that the events still end up hitting my on-prem instance (D'oh if i'm trying to generate thousands/millions in seconds). HOWEVER this is crucial if I want to compare pre-ingestion to post-ingestion. Well, to avoid hitting Sentry/localhost:9000, I could maybe run this experiment inside of a Network where all http requests gets routed through a Proxy which can also read the request payloads, but then terminate the request from reaching the on-prem Sentry?  

THOUGHT - sentry sdk 'Custom Transport Layers'? This would be really easy if I could just have sentry sdk send the events to a URL of my choice, negating the need for a on-prem-Sentry. Could someone help me with this? See comments in deprecated/dump-request.go, I could send sentry sdk events to the router defined in there

example payload structure from a sentry sdk event:  
![payload-structure](./payload-structure.png)

## Versions
tested on ubuntu 18.04 LTS

go version go1.12.9 linux/amd64

sentry-sdk==0.14.2

## Install
```
virtualenv -p /usr/bin/python3 .virtualenv  
source .virtualenv/bin/activate  
pip3 install requirements.txt
```

download gor executable and put to cwd  
https://github.com/buger/goreplay/releases/tag/v1.0.0

```
go get github.com/buger/goreplay/proto  
go get github.com/buger/jsonparser
```

run https://github.com/getsentry/onpremise, it defaults to localhost:9000
visit http://localhost:9000 and get a dsn key, put it in a new .env file so app.py reads from that

```
go build middleware.go
```

## Run
1. `sudo ./gor --input-raw :9000 --middleware "./middleware" --output-stdout`
2. `python3 app.py`

^ see the debug log statement in your terminal, it logs the platform property of the event (i.e. event.platform, should read "python")  
^ NEXT - log the entire payload / persist it somewhere for ML

## Reference & Troubleshooting

#### Sentry
https://github.com/getsentry/sentry-python  
https://github.com/getsentry/sentry-go  
https://github.com/getsentry/onpremise  
Borrowed code from https://github.com/getsentry/gor-middleware/blob/master/auth.go

#### buger's goreplay
https://github.com/buger/jsonparser

I used this as my 'middleware.go' and removed what I didn't need:  
https://github.com/buger/goreplay/blob/master/examples/middleware/token_modifier.go

About the middleware technique  
https://github.com/buger/goreplay/tree/master/middleware

#### other
This 'DumpRequest' (deprecated/dump-request.go) would be perfect if I could make sentry_sdk send events to a URL of my choosing. Downside is the events would never reach my on-prem Sentry. Maybe support both techniques in this repo:  
https://rominirani.com/golang-tip-capturing-http-client-requests-incoming-and-outgoing-ef7fcdf87113

https://golang.org/pkg/net/http/#Request  
https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body using encoding/json instead of buger/jsonparser  

gor file-server 8000

// basic gor usage, without a middleware like middleware.go  
sudo ./gor --input-raw :8000 --output-stdout

## TODO
.mod this