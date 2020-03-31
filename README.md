# event-maker
## Goal
Test data automation.

Sending diverse events from multiple sdk types to a Sentry Organization on a regular basis.

Keep 1 app running instead of 1 app per platform.


## What's Happening
<img src="./img/workflow-diagram.jpeg" width="450" height="300">  

STEP1 - Sentry sdk's send events to the API defined in /flask/server.py. It acts like a proxy that intercept the events before they hit Sentry. It saves copies of them in a database. This is useful because apps w/ sdk's do not have to stay running on a scheduled job to keep creating more errors and events. Events are instead saved in a database for replaying in the future.

STEP2 - Events do not have to be created because they're alread stored in a database. Load the events from the database and send them to Sentry. This can run on a scheduled job. Sentry thinks they're coming from live apps.

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
3. `make` runs Flask server
4. `python app.py`
5. `psql` or `/event-bytea` to load the event again
or
6. `localhost:9000` to see your Sentry onprem event, if you used forwarding.

or
7. just load event `/event-bytea` and forward to Sentry

Workflow:  
`python app.py` sdk sends event to the intercetpor.

The `DSN` that you use in your `app.py` determine what the proxy will do. They are mapped to different endpoints in `flask/app.py`.

`localhost:3001/impersonator` will load an event from the database and forward it (by http) to your Sentry instance.

'STEP1' endpoints require an sdk that sends an event to them

'STEP2' endpoints you can hit yourself from Postman

you may have to `sudo service postgresql stop` to free up 5432 on your machine

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

https://rominirani.com/golang-tip-capturing-http-client-requests-incoming-and-outgoing-ef7fcdf87113
https://golang.org/pkg/net/http/#Request  
https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body using encoding/json instead of buger/jsonparser  
gor file-server 8000

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
```

replaying the payload many times. grpc

MemoryView  
https://www.postgresql.org/message-id/25EDB20679154BDBB3CBBD335184E1D7%40frank  
https://www.postgresql.org/message-id/C2C12FD0FCE64CE8BB77765A526D3C73%40frank  

"Q. How to save a instance of a Class to the DB?"
"A. You can't store the object itself in the DB. What you do is to store the data from the object and reconstruct it later."
https://stackoverflow.com/questions/2047814/is-it-possible-to-store-python-class-objects-in-sqlite

Troubleshoot - compare len(bytes) on the way in as when it came out...

{\"exception\": 
    {
        \"values\":
         [{\"stacktrace\": 
            {\"frames\": []},
             \"type\": \"Exception\", \"value\": \"ooooooo\", \"module\": \"exceptions\", \"mechanism\": null}]
    }, 
    \"server_name\": \"pop-os\", \"level\": \"error\",
    \"event_id\": \"d00a16bf0c2a485283c82c2f962835bb\",
     "timestamp\": \"2020-03-30T03:47:37.000588Z\", 
     \"extra\": {\"sys.argv\": [\"app.py\"]}, \"modules\": {\"pandocfilters\": \"1.4.2\", \"ipython-genutils\": \"0.2.0\", \"oauth\": \"1.0.1\", \"attrs\": \"19.3.0\", \"pyparsing\": \"2.4.0\", \"keyrings.alt\": \"3.0\", \"jupyterlab-launcher\": \"0.11.2\", \"lazr.uri\": \"1.0.3\", \"flask\": \"1.1.1\", \"send2trash\": \"1.5.0\", \"dotenv\": \"0.0.5\", \"itsdangerous\": \"1.1.0\", \"prometheus-client\": \"0.7.1\", \"pathlib2\": \"2.3.5\", \"backports.shutil-get-terminal-size\": \"1.0.0\", \"python\": \"2.7.17\", \"secretstorage\": \"2.3.1\", \"markupsafe\": \"1.1.1\", \"jinja2\": \"2.11.1\", \"httplib2\": \"0.9.2\", \"bleach\": \"3.1.1\", \"decorator\": \"4.4.2\", \"contextlib2\": \"0.6.0.post1\", \"jupyter-client\": \"5.3.4\", \"wadllib\": \"1.3.2\", \"psutil\": \"5.4.2\", \"cycler\": \"0.10.0\", \"jsonschema\": \"3.2.0\", \"ipywidgets\": \"7.5.1\", \"kiwisolver\": \"1.1.0\", \"sentry-sdk\": \"0.14.2\", \"ptyprocess\": \"0.6.0\", \"importlib-metadata\": \"1.5.0\", \"qtpy\": \"1.9.0\", \"werkzeug\": \"1.0.0\", \"qtconsole\": \"4.7.0\", \"olefile\": \"0.45.1\", \"entrypoints\": \"0.3\", \"blinker\": \"1.4\", \"gunicorn\": \"19.10.0\", \"matplotlib\": \"2.2.4\", \"ipython\": \"5.9.0\", \"zipp\": \"1.2.0\", \"pickleshare\": \"0.7.5\", \"mistune\": \"0.8.4\", \"nbformat\": \"4.4.0\", \"pyxdg\": \"0.25\", \"wcwidth\": \"0.1.8\", \"wsgiref\": \"0.1.2\", \"traitlets\": \"4.3.3\", \"terminado\": \"0.8.3\", \"requests\": \"2.18.4\", \"defusedxml\": \"0.6.0\", \"simplegeneric\": \"0.8.1\", \"pillow\": \"5.1.0\", \"asn1crypto\": \"0.24.0\", \"pygobject\": \"3.26.1\", \"pygments\": \"2.5.2\", \"jupyter-console\": \"5.2.0\", \"prompt-toolkit\": \"1.0.18\", \"pexpect\": \"4.8.0\", \"backports-abc\": \"0.5\", \"powerline-status\": \"2.6\", \"typing\": \"3.7.4.1\", \"python-dotenv\": \"0.12.0\", \"testpath\": \"0.4.4\", \"certifi\": \"2018.1.18\", \"numpy\": \"1.16.4\", \"pyzmq\": \"19.0.0\", \"sqlalchemy\": \"1.3.15\", \"simplejson\": \"3.13.2\", \"widgetsnbextension\": \"3.5.1\", \"subprocess32\": \"3.5.4\", \"powerline-shell\": \"0.7.0\", \"pytz\": \"2019.1\", \"jupyter-core\": \"4.6.3\", \"functools32\": \"3.2.3.post2\", \"python-dateutil\": \"2.8.1\", \"jupyterlab\": \"0.33.12\", \"pycrypto\": \"2.6.1\", \"pyrsistent\": \"0.15.7\", \"chardet\": \"3.0.4\", \"setuptools\": \"44.0.0\", \"flask-cors\": \"3.0.8\", \"configobj\": \"5.0.6\", \"ipykernel\": \"4.10.1\", \"zope.interface\": \"4.3.2\", \"backports.functools-lru-cache\": \"1.5\", \"singledispatch\": \"3.4.0.3\", \"pip\": \"20.0.2\", \"configparser\": \"4.0.2\", \"cryptography\": \"2.1.4\", \"six\": \"1.14.0\", \"click\": \"7.1.1\", \"nbconvert\": \"5.6.1\", \"lazr.restfulclient\": \"0.13.5\", \"webencodings\": \"0.5.1\", \"wheel\": \"0.30.0\", \"tornado\": \"5.1.1\", \"urllib3\": \"1.22\", \"notebook\": \"5.7.8\", \"ipaddress\": \"1.0.23\", \"launchpadlib\": \"1.10.6\", \"argparse\": \"1.2.1\", \"jupyter\": \"1.0.0\", \"bzr\": \"2.8.0.dev1\", \"psycopg2-binary\": \"2.8.4\", \"enum34\": \"1.1.9\", \"futures\": \"3.3.0\", \"keyring\": \"10.6.0\", \"pyopenssl\": \"17.5.0\", \"idna\": \"2.6\", \"nltk\": \"3.4.3\", \"scandir\": \"1.10.0\"}, \"contexts\": {\"runtime\": {\"version\": \"2.7.17\", \"name\": \"CPython\", \"build\": \"2.7.17 (default, Nov  7 2019, 10:07:09) \\n[GCC 7.4.0]\"}}, \"platform\": \"python\", \"breadcrumbs\": [], \"sdk\": {\"version\": \"0.14.2\", \"name\": \"sentry.python\", \"packages\": [{\"version\": \"0.14.2\", \"name\": \"pypi:sentry-sdk\"}], \"integrations\": [\"argv\", \"atexit\", \"dedupe\", \"excepthook\", \"logging\", \"modules\", \"stdlib\", \"threading\"]
     }

}
