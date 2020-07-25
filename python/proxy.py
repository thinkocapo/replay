import os
import datetime
from dotenv import load_dotenv
from flask import Flask, request, json, abort
from flask_cors import CORS
import json
from services import compress_gzip, decompress_gzip, get_event_type
import sqlite3
import string # ?
import urllib3
import uuid
load_dotenv()
http = urllib3.PoolManager()

app = Flask(__name__)

app.run(threaded=True)
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


""" This is only for using the proxy to forward events directly to Sentry and NOT save them in your database
If you're not using this, you can ignore it
Must pass auth key in URL (not request headers) or else 403 CSRF error from Sentry
AM Transactions can't be sent to any self-hosted Sentry instance as of 10.0.0 05/30/2020 
"""
def sentryUrl(DSN):
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

SQLITE = os.getenv('SQLITE')
database = SQLITE or os.getcwd() + "/sqlite.db"
print("> database", database)

with sqlite3.connect(database) as db:
    cursor = db.cursor()
    cursor.execute(""" CREATE TABLE IF NOT EXISTS events (
                                            id integer PRIMARY KEY,
                                            platform text,
                                            type text,
                                            data BLOB,
                                            headers BLOB
                                        ); """)
    cursor.close()

# MODIFIED_DSN_FORWARD - Intercepts the payload sent by sentry_sdk in event.py, and then sends it to a Sentry instance
@app.route('/api/2/store/', methods=['POST'])
def forward():
    print('> FORWARD')

    # TODO exception.platform may have been available, as well as exception.sdk https://github.com/thinkocapo/undertaker/issues/48
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

import sentry_sdk
# sentry_sdk.init(
#     dsn="https://f5227a4c11874545948bd39dd95ed7b4@o87286.ingest.sentry.io/5314428",
#     release='0.0.1'    
# )

@app.route('/api/5/envelope/', methods=['POST'])
def save_mobile_envelope():
    print('\n ******* ENVELOPE ******* ')

    print('> /api/5/envelope ')

    print('> type(request.data)', type(request.data)) # is STRING
    print('> type(request_headers)', type(request.headers))

    # for header in request.headers:
    #     print(header)

    event_platform = ''
    event_type = ''
    request_headers = {}
    user_agent = request.headers.get('User-Agent').lower()
    body = ''


    print('\nXXXXXXXXXXX')
    print(decompress_gzip(request.data))
    print('\nYYYYYYYYYYY')
    # print(json.dumps(json.loads(decompress_gzip(request.data)),indent=2))

    event_platform = 'android'
    event_type = get_event_type(request.data, "android")
    print('> event_type11111', event_type)
    
    for key in ['X-Sentry-Auth', 'Content-Length','User-Agent','Connection','Content-Encoding','X-Forwarded-Proto','Host','Accept','X-Forwarded-For', 'Content-Type', 'Accept-Encoding']:
        request_headers[key] = request.headers.get(key)
    # print(json.dumps(request_headers,indent=2))

    # print(json.dumps(json.loads(decompress_gzip(request.data)),indent=2))
    # or for sessions:
    print(json.dumps(decompress_gzip(request.data),indent=2))
    
    # TODO is not hte right type yet...
    body = decompress_gzip(request.data)

    insert_query = ''' INSERT INTO events(platform,type,body,headers)
              VALUES(?,?,?,?) '''
    record = (event_platform, event_type, body, json.dumps(request_headers))
    try:
        with sqlite3.connect(database) as db:
            cursor = db.cursor()
            cursor.execute(insert_query, record)
            print('> SQLITE ID', cursor.lastrowid)
            cursor.close()
            return str(cursor.lastrowid)
    except Exception as err:
        print("LOCAL EXCEPTION", err)

    print('> SAVING /api/5/envelope END')

    return 'SUCCESS'

# MODIFIED_DSN_SAVE MOBILE - Intercepts event from sentry sdk and saves them to Sqlite DB. No forward of event to your Sentry instance.
@app.route('/api/5/store/', methods=['POST'])
def save_mobile():
    print('> /api/5/store')

    print('> type(request.data)', type(request.data)) # STRING
    print('> type(request_headers)', type(request.headers))

    event_platform = ''
    event_type = ''
    request_headers = {}
    user_agent = request.headers.get('User-Agent').lower()
    body = ''

    event_platform = 'android'
    event_type = get_event_type(request.data, "android")
    print('> event_type', event_type)
    
    for key in ['X-Sentry-Auth', 'Content-Length','User-Agent','Connection','Content-Encoding','X-Forwarded-Proto','Host','Accept','X-Forwarded-For']:
        request_headers[key] = request.headers.get(key)
    # print(json.dumps(request_headers,indent=2))
    # print(json.dumps(json.loads(decompress_gzip(request.data)),indent=2))
    body = decompress_gzip(request.data)

    insert_query = ''' INSERT INTO events(platform,type,body,headers)
              VALUES(?,?,?,?) '''
    record = (event_platform, event_type, body, json.dumps(request_headers))
    try:
        with sqlite3.connect(database) as db:
            cursor = db.cursor()
            cursor.execute(insert_query, record)
            print('> SQLITE ID', cursor.lastrowid)
            cursor.close()
            return str(cursor.lastrowid)
    except Exception as err:
        print("LOCAL EXCEPTION", err)

    print('> SAVING /api/5/store END')
    return 'SUCCESS'


# MODIFIED_DSN_SAVE - Intercepts event from sentry sdk and saves them to Sqlite DB. No forward of event to your Sentry instance.
@app.route('/api/3/store/', methods=['POST'])
def save():
    print('> SAVING')

    event_platform = ''
    event_type = ''
    request_headers = {}
    user_agent = request.headers.get('User-Agent').lower()
    body = ''

    if 'python' in user_agent:

        event_platform = 'python'
        event_type = get_event_type(request.data, "python")
        print('> PYTHON', event_type)

        for key in ['Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent', 'X-Sentry-Auth']:
            request_headers[key] = request.headers.get(key)

        body = decompress_gzip(request.data)

    if 'mozilla' in user_agent or 'chrome' in user_agent or 'safari' in user_agent:

        event_platform = 'javascript'
        event_type = get_event_type(request.data, "javascript")
        print('> JAVASCRIPT ', event_type)

        for key in ['Accept-Encoding','Content-Length','Content-Type','User-Agent']:
            request_headers[key] = request.headers.get(key)

        body = request.data

    insert_query = ''' INSERT INTO events(platform,type,body,headers)
              VALUES(?,?,?,?) '''
    record = (event_platform, event_type, body, json.dumps(request_headers))
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

    # print('> type(request.data)', type(request.data))
    # print('> type(request_headers)', type(request.headers))
    # for header in request.headers.to_wsgi_list():
    #     print(header)
    # print(json.dumps(json.loads(decompress_gzip(request.data)),indent=2))
    # json.dumps(json.loads(request.data),indent=2)

    request_headers = {}
    for key in ['Host','Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent']:
        request_headers[key] = request.headers.get(key)

    insert_query = ''' INSERT INTO events(platform,type,body,headers)
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
