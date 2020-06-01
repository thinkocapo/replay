import os
import datetime
from dotenv import load_dotenv
from flask import Flask, request, json, abort
from flask_cors import CORS
import json
# import sentry_sdk
# from sentry_sdk.integrations.flask import FlaskIntegration
from services import compress_gzip, decompress_gzip, get_event_type
import sqlite3
import string # ?
import urllib3
import uuid
load_dotenv()
http = urllib3.PoolManager()

app = Flask(__name__)
# app.run(ssl_context='adhoc') # flask run --cert=adhoc
CORS(app)

print("""
                               Welcome To The
  _   _   _   _   ____    _____   ____    _____      _      _  __  _____   ____  
 | | | | | \ | | |  _ \  | ____| |  _ \  |_   _|    / \    | |/ / | ____| |  _ \ 
 | | | | |  \| | | | | | |  _|   | |_) |   | |     / _ \   | ' /  |  _|   | |_) |
 | |_| | | |\  | | |_| | | |___  |  _ <    | |    / ___ \  | . \  | |___  |  _ < 
  \___/  |_| \_| |____/  |_____| |_| \_\   |_|   /_/   \_\ |_|\_\ |_____| |_| \_\
                                                                                 
""")



SENTRY=''

# Must pass auth key in URL (not request headers) or else 403 CSRF error from Sentry
# AM Transactions can't be sent to any self-hosted SEntry instance as of 05/30/2020 
# https://github.com/getsentry/sentry/releases
def sentryUrl(DSN):
    print('> sentryUrl')

    if ("@localhost:" in DSN):
        KEY = DSN.split('@')[0][7:]
        # assumes single-digit projectId right now
        PROJECT_ID= DSN[-1:]
        HOST = 'localhost:9000'
        return "http://%s/api/%s/store/?sentry_key=%s&sentry_version=7" % (HOST, PROJECT_ID, KEY)
    if ("ingest.sentry.io" in DSN):
        KEY = DSN.split('@')[0][8:] # 8 because of 's' in 'https'
        HOST = DSN.split('@')[1].split('/')[0]
        PROJECT_ID = DSN.split('@')[1].split('/')[1] 
        return "https://%s/api/%s/store/?sentry_key=%s&sentry_version=7" % (HOST, PROJECT_ID, KEY)
        # MODIFIED_DSN_FORWARD used a dsn of "http://0d52d5f4e8a64f5ab2edce50d88a7626@o87286.ingest.sentry.io/1428657" in thinkocapo/react to call:
        # return "https://o87286.ingest.sentry.io/api/1428657/store/?sentry_key=0d52d5f4e8a64f5ab2edce50d88a7626&sentry_version=7" # will-frontend-react in SAAS
        # it ^ reached SaaS sentry.io


# DATABASE - Must be full absolute path to sqlite database file
# sqlite.db will get created if doesn't exist
SQLITE = os.getenv('SQLITE')
database = SQLITE or os.getcwd() + "/sqlite.db"
print("> database", database)

with sqlite3.connect(database) as db:
    # TODO platform, eventType instead of name, type. for now, re-purpose 'name' as 'platform' and 'type' as 'eventType'
    cursor = db.cursor()
    cursor.execute(""" CREATE TABLE IF NOT EXISTS events (
                                            id integer PRIMARY KEY,
                                            name text,
                                            type text,
                                            data BLOB,
                                            headers BLOB
                                        ); """)
    cursor.close()

########################  STEP 1  #########################

# MODIFIED_DSN_FORWARD - Intercepts the payload sent by sentry_sdk in event.py, and then sends it to a Sentry instance
@app.route('/api/2/store/', methods=['POST'])
def forward():
    print('> FORWARD')

    # TODO https://github.com/thinkocapo/undertaker/issues/48
    def make(headers):
        request_headers = {}
        user_agent = request.headers.get('User-Agent').lower()
        if 'python' in user_agent:
            print('> Python error')
            for key in ['Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent']:
                request_headers[key] = request.headers.get(key)
                # SENTRY = sentryUrl(os.getenv('DSN_PYTHON'))
                # SENTRY = sentryUrl(os.getenv('DSN_PYTHON_SAAS'))
                SENTRY = sentryUrl(os.getenv('DSN_PYTHONEAT_SAAS'))
        if 'mozilla' in user_agent or 'chrome' in user_agent or 'safari' in user_agent:
            print('> Javascript error')
            for key in ['Accept-Encoding','Content-Length','Content-Type','User-Agent']:
                request_headers[key] = request.headers.get(key)
                # SENTRY = sentryUrl(os.getenv('DSN_REACT'))
                SENTRY = sentryUrl(os.getenv('DSN_REACT_SAAS'))
        return request_headers, SENTRY

    request_headers, SENTRY = make(request.headers)
    print('> SENTRY url for store endpoint', SENTRY)

    try:
        print('> type(request.data)', type(request.data))
        print('> type(request_headers)', type(request_headers))

        response = http.request(
            "POST", str(SENTRY), body=request.data, headers=request_headers 
        )

        print('> nothing saved to sqlite database')
        return 'success'
    except Exception as err:
        print('LOCAL EXCEPTION', err)

# MODIFIED_DSN_SAVE - Intercepts event from sentry sdk and saves them to Sqlite DB. No forward of event to your Sentry instance.
@app.route('/api/3/store/', methods=['POST'])
def save():
    print('> SAVING')

    # print('> type(request.data)', type(request.data))
    # print('> type(request_headers)', type(request.headers))
    # for header in request.headers.to_wsgi_list():
    #     print(header)
    # print(json.dumps(json.loads(decompress_gzip(request.data)),indent=2))
    # json.dumps(json.loads(request.data),indent=2)

    event_platform = ''
    event_type = ''
    request_headers = {}
    user_agent = request.headers.get('User-Agent').lower()
    
    data = ''
    if 'python' in user_agent:

        event_platform = 'python'
        event_type = get_event_type(request.data, "python")
        print('> PYTHON', event_type)

        for key in ['Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent', 'X-Sentry-Auth']:
            request_headers[key] = request.headers.get(key)

        data = decompress_gzip(request.data)

    if 'mozilla' in user_agent or 'chrome' in user_agent or 'safari' in user_agent:

        event_platform = 'javascript'
        event_type = get_event_type(request.data, "javascript")
        print('> JAVASCRIPT ', event_type)

        for key in ['Accept-Encoding','Content-Length','Content-Type','User-Agent']:
            request_headers[key] = request.headers.get(key)

        data = request.data

    insert_query = ''' INSERT INTO events(name,type,data,headers)
              VALUES(?,?,?,?) '''
    record = (event_platform, event_type, data, json.dumps(request_headers))
    try:
        with sqlite3.connect(database) as db:
            cursor = db.cursor()
            cursor.execute(insert_query, record)
            print('> SQLITE ID', cursor.lastrowid)
            cursor.close()
            return str(cursor.lastrowid)
    except Exception as err:
        print("LOCAL EXCEPTION", err)

# MODIFIED_DSN_SAVE_AND_FORWARD - this has been out of date since proxy.py started supporting Transactions in /api/2/store and /api/3/store endpoints
@app.route('/api/4/store/', methods=['POST'])
def save_and_forward():

    request_headers = {}
    for key in ['Host','Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent']:
        request_headers[key] = request.headers.get(key)

    insert_query = ''' INSERT INTO events(name,type,data,headers)
              VALUES(?,?,?,?) '''
    record = ('python', 'example', request.data, json.dumps(request_headers)) # type(json.dumps(request_headers)) <type 'str'>

    try:
        with sqlite3.connect(database) as db:
            cursor = db.cursor()
            cursor.execute(insert_query, record)
            print('> sqlite3 ID', cursor.lastrowid)
            cursor.close()
    except Exception as err:
        print("LOCAL EXCEPTION SAVE", err)

    try:
        response = http.request(
            "POST", str(SENTRY), body=request.data, headers=request_headers 
        )
        return 'success'
    except Exception as err:
        print('LOCAL EXCEPTION FORWARD', err)
