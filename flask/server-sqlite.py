import os
import datetime
from dotenv import load_dotenv
from flask import Flask, request, json, abort
from flask_cors import CORS
import gzip
from gzip import GzipFile
import io
import json
# import sentry_sdk
# from sentry_sdk.integrations.flask import FlaskIntegration
from six import BytesIO
import sqlite3
import string # ?
import urllib3
import uuid
load_dotenv()

app = Flask(__name__)
CORS(app)

http = urllib3.PoolManager()

print("""
                               Welcome To The
  _   _   _   _   ____    _____   ____    _____      _      _  __  _____   ____  
 | | | | | \ | | |  _ \  | ____| |  _ \  |_   _|    / \    | |/ / | ____| |  _ \ 
 | | | | |  \| | | | | | |  _|   | |_) |   | |     / _ \   | ' /  |  _|   | |_) |
 | |_| | | |\  | | |_| | | |___  |  _ <    | |    / ___ \  | . \  | |___  |  _ < 
  \___/  |_| \_| |____/  |_____| |_| \_\   |_|   /_/   \_\ |_|\_\ |_____| |_| \_\
                                                                                 
""")

# SENTRY - Must pass auth key in URL (not request headers) or else 403 CSRF error from Sentry
SENTRY ="http://localhost:9000/api/2/store/?sentry_key=09aa0d909232457a8a6dfff118bac658&sentry_version=7"

# DATABASE - Must be full absolute path to sqlite database file
SQLITE = os.getenv('SQLITE')
# sqlite.db will get created if doesn't exist
database = SQLITE or os.getcwd() + "/sqlite.db"
print(" > database", database)
with sqlite3.connect(database) as conn:
    cur = conn.cursor()
    cur.execute(""" CREATE TABLE IF NOT EXISTS events (
                                            id integer PRIMARY KEY,
                                            name text,
                                            type text,
                                            data BLOB,
                                            headers BLOB
                                        ); """)
    cur.close()

# Functions from getsentry/sentry-python
def decompress_gzip(bytes_encoded_data):
    try:
        fp = BytesIO(bytes_encoded_data)
        try:
            f = GzipFile(fileobj=fp)
            return f.read().decode("utf-8")
        finally:
            f.close()
    except Exception as e:
        raise e

def compress_gzip(dict_body):
    try:
        body = io.BytesIO()
        with gzip.GzipFile(fileobj=body, mode="w") as f:
            f.write(json.dumps(dict_body, allow_nan=False).encode("utf-8"))
    except Exception as e:
        raise e
    return body

########################  STEP 1  #########################

# MODIFIED_DSN_FORWARD - Intercepts the payload sent by sentry_sdk in app.py, and then sends it to a Sentry instance
@app.route('/api/2/store/', methods=['POST'])
def forward():

    request_headers = {}
    for key in ['Host','Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent']:
        request_headers[key] = request.headers.get(key)
    
    try:
        print('> type(request.data)', type(request.data))
        print('> type(request_headers)', type(request_headers))

        response = http.request(
            "POST", str(SENTRY), body=request.data, headers=request_headers 
        )
        # print("> RESPONSE and event_id %s" % (response.status, response.data))
        return 'success'
    except Exception as err:
        print('LOCAL EXCEPTION', err)

# MODIFIED_DSN_SAVE - Intercepts event from sentry sdk and saves them to Sqlite DB. No forward of event to your Sentry instance.
@app.route('/api/3/store/', methods=['POST'])
def save():
    print('> SAVING')
    # type(request.data) is <class 'bytes'>
    request_headers = {}
    for key in ['Host','Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent']:
        request_headers[key] = request.headers.get(key)

    insert_query = ''' INSERT INTO events(name,type,data,headers)
              VALUES(?,?,?,?) '''
    record = ('python1', 'python', request.data, json.dumps(request_headers))
   
    try:
        with sqlite3.connect(database) as conn:
            cur = conn.cursor()
            cur.execute(insert_query, record)
            print('sqlite3 row ID', cur.lastrowid)
            cur.close()
            return str(cur.lastrowid)
    except Exception as err:
        print("LOCAL EXCEPTION", err)

# MODIFIED_DSN_SAVE_AND_FORWARD
@app.route('/api/4/store/', methods=['POST'])
def save_and_forward():

    request_headers = {}
    for key in ['Host','Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent']:
        request_headers[key] = request.headers.get(key)

    # insert_query = """ INSERT INTO events (type, name, data, headers) VALUES (%s,%s,%s,%s)"""
    insert_query = ''' INSERT INTO events(name,type,data,headers)
              VALUES(?,?,?,?) '''
    record = ('python', 'example', request.data, json.dumps(request_headers)) # type(json.dumps(request_headers)) <type 'str'>

    try:
        with sqlite3.connect(database) as conn:
            cur = conn.cursor()
            cur.execute(insert_query, record)
            print('> sqlite3 ID', cur.lastrowid)
            cur.close()
    except Exception as err:
        print("LOCAL EXCEPTION SAVE", err)

    try:
        response = http.request(
            "POST", str(SENTRY), body=request.data, headers=request_headers 
        )
        print("> RESPONSE and event_id %s" % (response.status, response.data))
        return 'response not read by client sdk'
    except Exception as err:
        print('LOCAL EXCEPTION FORWARD', err)

# see event-to-sentry.py
########################  STEP 2  #########################

# Loads a saved event's payload+headers from database and forwards to Sentry instance 
# if no pk ID is provided then query selects most recent event
# @app.route('/load-and-forward', defaults={'pk':0}, methods=['GET'])
# @app.route('/load-and-forward/<pk>', methods=['GET'])
# def load_and_forward(pk):
    
#     if pk==0:
#         query = "SELECT * FROM events ORDER BY pk DESC LIMIT 1;"
#     else:
#         query = "SELECT * FROM events WHERE pk={};".format(pk)

#     with sqlite3.connect(database) as conn:
#         cur = conn.cursor()
#         cur.execute("SELECT * FROM events ORDER BY id DESC LIMIT 1;")
#         rows = cur.fetchall()
#         row = rows[0]
#         row = list(row)
#         # 'bytes' not 'buffer' like in db_prep.py
#         body_bytes_buffer = row[3] # not row_proxy.data, because sqlite returns tuple (not row_proxy)
#         request_headers = json.loads(row[4])
#         # print('\n type(body_bytes_buffer)', type(body_bytes_buffer))

#     # call it 'bytes_buffer_body'
#     # update event_id/timestamp so Sentry will accept the event again
#     json_body = decompress_gzip(body_bytes_buffer)
#     dict_body = json.loads(json_body)
#     dict_body['event_id'] = uuid.uuid4().hex
#     dict_body['timestamp'] = datetime.datetime.utcnow().isoformat() + 'Z'
#     print('> event_id', dict_body['event_id'])
#     print('> timestamp', dict_body['timestamp'])
#     bytes_io_body = compress_gzip(dict_body)
        
#     try:
#         # print('type(request_headers)', type(request_headers))
#         # print('type(bytes_io_body)', type(bytes_io_body))
#         # print('type(bytes_io_body.getvalue())', type(bytes_io_body.getvalue()))

#         response = http.request(
#             "POST", str(SENTRY), body=bytes_io_body.getvalue(), headers=request_headers
#         )
#     except Exception as err:
#         print('LOCAL EXCEPTION', err)

#     return("> FINISH")

