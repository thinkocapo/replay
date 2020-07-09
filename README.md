<!-- ![The Undertaker](./img/undertaker-1.png) -->
# The Undertaker
The Undertaker a.k.a. Replay is an event traffic replay service.

<img src="./img/undertaker-4.jpeg" width="450" height="300">  

## Why?  
Stop maintaing +10 different platform SDK's in GCP sending events all the time. Rather, run a single program `event-to-sentry.go` on a cronjob to send all those events for you. It's free. Great for automating test data.

<img src="./img/event-maker-slide-2.001.png" width="450" height="300">  

**STEP 1**  
SDK's create events and `python/proxy.py` intercepts "undertakes" the events on their way to Sentry and saves them in sqlite

**STEP 2**  
`event-to-sentry.go` loads the events from sqlite and relpays them to Sentry (no sdk's used in this step)

## Setup

1. Enter your DSN's in `.env`  
```
// for the Tool Store data set (go build -o bin/event-to-sentry-toolstore *.go)
DSN_JAVASCRIPT_SAAS=
DSN_PYTHON_SAAS=

or

// for the Gateway/Microservices/Celery dataset (go build -o bin/event-to-sentry-tracing-example *.go)
DSN_PYTHON_GATEWAY=
DSN_PYTHON_DJANGO=
DSN_PYTHON_CELERY=

// set this here as your default or pass it at runtime using --db=
SQLITE=
```

2. `pip3 install -r ./python/requirements.txt` for the proxy  
3.

```
go build -o bin/event-to-sentry-<name> *.go

// for Tool Store data set (javascrip, pythont, errors+transactions)
go build -o bin/event-to-sentry-toolstore

// for Gateway/Microservices/Celery dataset (python transactions)
go build -o bin/event-to-sentry-tracing-example *.go
```

Note - Transactions are not supported if using DSN's from `getsentry/onpremise` as of 07/08/20

## Run
```
./bin/event-to-sentry-<name>
./bin/event-to-sentry
./bin/event-to-sentry --id=<id>
./bin/event-to-sentry --id=<id> -i
./bin/event-to-sentry --all
```
or use `--js` `--py` to pass DSN's when running the executable
```
./bin/event-to-sentry --all --db=am-transactions-timeout-sqlite.db
./bin/event-to-sentry --all --db=<path_to_.db> --js=<javascripti_DSN> --py=<python_DSN>
```

See your events in Sentry

## Proxy (optional)
Use the proxy if you want to create your own data set 

1. Get your proxy running
```
make proxy
```

2. Modify your app's DSN so it will point to the proxy. See `python/event.py` for how to do this.

3. Create errors in your app, so the events get sent to the proxy.

4. Check your events saved to the database
`python3 test/db.py` or make testdb

Note - You can expose the proxy's port 3001 via ngrok, if your apps are in a VPC/network that you can't run the proxy inside of. 
1. `ngrok http 3001`
2. put the ngrok address in your app's DSN like:  
`SENTRY_DSN=https://1f2d7bf845114ba6a5ba19ee07db6800@5b286dac3e72.ngrok.io/3`
3. now your events will send to the proxy

## Cronjobs
Macbook's cronjob manager for sending events in the background while you work
```
# crontab -l, to list cronjobs
# crontab -e to open crontab manager

# every minute
1-59 * * * * cd /<path>/<to>/undertaker/ && ./event-to-sentry

# every minute, every day of the week M-F
# * * * * 1-5 cd /<path>/<to>/undertaker/ && ./event-to-sentry-<name> --all

# every 5 minutes
*/5 * * * 1-5 cd /<path>/<to>/undertaker/ && ./event-to-sentry-<name> --all
```

https://crontab.guru/

## Notes

#### database
`python3 test/db.py` shows total event count and most recently saved event.  
`python3 test/db.py 5` gets the 5th event  
`python3 test/dby.py 5 -b` gets the 5th event and prints its body  

#### gotcha's
The timestamp from `go run event-to-sentry.go` is sometimes earlier than today's date and time 
Use python3 or else else `getvalue()` in `python/event-to-sentry.py` returns wrong data type

#### other
Borrowed code from: getsentry/sentry-python, getsentry/sentry-go, goreplay, getsentry/gor-middleware

https://develop.sentry.dev/sdk/store for info on the Sentry store endpoint

https://develop.sentry.dev/sdk/event-payloads/ for what a sdk event looks like. Here's [/img/example-payload.png](./img/example-payload.png) from javascript

6 events in the am-transactions-sqlite.db was 57kb
19 events tracing-example was 92kb

tested on: ubuntu 18.04 LTS, go 1.12.9 linux/amd64, sentry-sdk 0.14.2, flask Python 3.6.9

`python-dotenv` vs `dotenv`

## Todo
- Mobile android errors/crashes/sessions
- update tracing-example's endpoint names. www.toolstoredmeo.com instead of gcp url

















sentry-cli releases -o testorg-az new -p proxy 0.0.2

sentry-cli releases -o testorg-az -p proxy set-commits --local 0.0.2






