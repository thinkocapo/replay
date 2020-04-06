import os
import datetime
from flask import Flask, request, json, abort
from flask_cors import CORS
import gzip
import uuid
import json
import requests
import sentry_sdk
from sentry_sdk.integrations.flask import FlaskIntegration
import sqlalchemy
from sqlalchemy import create_engine
import io
from six import BytesIO
from gzip import GzipFile
import urllib3
http = urllib3.PoolManager()

import psycopg2
import string
import psycopg2.extras

# Must pass auth key in URL (not request headers) or else 403 CSRF error from Sentry
SENTRY_API_STORE_ONPREMISE ="http://localhost:9000/api/2/store/?sentry_key=759bf0ad07984bb3941e677b35a13d2c&sentry_version=7"

app = Flask(__name__)
CORS(app)

# Database
HOST='localhost'
DATABASE='postgres'
USERNAME='admin'
PASSWORD='admin'
db = create_engine('postgresql://' + USERNAME + ':' + PASSWORD + '@' + HOST + ':5432/' + DATABASE)

# sometimes needed in endpoint
# Database - set typecasting so psycopg2 returns bytea type as 'bytes' and not 'MemoryView'
# def bytea2bytes(value, cur):
#     m = psycopg2.BINARY(value, cur)
#     if m is not None:
#         return m.tobytes()
# BYTEA2BYTES = psycopg2.extensions.new_type(
#     psycopg2.BINARY.values, 'BYTEA2BYTES', bytea2bytes)
# psycopg2.extensions.register_type(BYTEA2BYTES)

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
        response = http.request(
            "POST", str(SENTRY_API_STORE_ONPREMISE), body=request.data, headers=request_headers 
        )

        print("%s RESPONSE and event_id %s" % (response.status, response.data))
        return 'success'
    except Exception as err:
        print('LOCAL EXCEPTION', err)

    return 'event was impersonated to Sentry'

# MODIFIED_DSN_SAVE - Intercepts event from sentry sdk and saves them to DB. No forward of event to your Sentry instance.
@app.route('/api/3/store/', methods=['POST'])
def save():

    request_headers = {}
    for key in ['Host','Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent']:
        request_headers[key] = request.headers.get(key)

    insert_query = """ INSERT INTO events (type, name, data, headers) VALUES (%s,%s,%s,%s)"""
    record = ('python', 'example', request.data, json.dumps(request_headers)) # type(json.dumps(request_headers)) <type 'str'>
    try:
        with db.connect() as conn:
            conn.execute(insert_query, record)
            conn.close()
        print("created event in postgres")
    except Exception as err:
        print("LOCAL EXCEPTION", err)

    return 'response not read by client sdk'

# MODIFIED_DSN_SAVE_AND_FORWARD
@app.route('/api/4/store/', methods=['POST'])
def save_and_forward():

    # Save
    request_headers = {}
    for key in ['Host','Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent']:
        request_headers[key] = request.headers.get(key)
    print('request_headers', request_headers)

    insert_query = """ INSERT INTO events (type, name, data, headers) VALUES (%s,%s,%s,%s)"""
    record = ('python', 'example', request.data, json.dumps(request_headers)) # type(json.dumps(request_headers)) <type 'str'>

    # TODO - TEST...the new try/except on db.connect here
    try:
        with db.connect() as conn:
            conn.execute(insert_query, record)
            conn.close()
    except Exception as err:
        print('LOCAL EXCEPTION', err)

    # Forward
    try:
        response = http.request(
            "POST", str(SENTRY_API_STORE_ONPREMISE), body=request.data, headers=request_headers 
        )
        print("%s RESPONSE and event_id %s" % (response.status, response.data))
        return 'response not read by client sdk'
    except Exception as err:
        print('LOCAL EXCEPTION', err)

########################  STEP 2  #########################

# Loads a saved event's payload+headers from database and forwards to Sentry instance 
# if no pk ID is provided then query selects most recent event
@app.route('/load-and-forward', defaults={'pk':0}, methods=['GET'])
@app.route('/load-and-forward/<pk>', methods=['GET'])
def load_and_forward(pk):
    # TODO 'If it's of class type memoryview then run this'
    # sometimes needed
    def bytea2bytes(value, cur):
        m = psycopg2.BINARY(value, cur)
        if m is not None:
            return m.tobytes()
    BYTEA2BYTES = psycopg2.extensions.new_type(
        psycopg2.BINARY.values, 'BYTEA2BYTES', bytea2bytes)
    psycopg2.extensions.register_type(BYTEA2BYTES)

    if pk==0:
        query = "SELECT * FROM events ORDER BY pk DESC LIMIT 1;"
    else:
        query = "SELECT * FROM events WHERE pk={};".format(pk)
    
    with db.connect() as conn:
        rows = conn.execute(query).fetchall()
        conn.close()
        # <class 'sqlalchemy.engine.result.RowProxy'
        row_proxy = rows[0]
    
    # check if it's of class type MemoryView or it's Bytes
    print('\nTYPE ', type(row_proxy.data))

    # row_proxy.data is <class bytes> so row_proxy.data is b'\x1f\x8b\
    json_body = decompress_gzip(row_proxy.data)

    # update event_id/timestamp so Sentry will accept the event again
    dict_body = json.loads(json_body)
    dict_body['event_id'] = uuid.uuid4().hex
    dict_body['timestamp'] = datetime.datetime.utcnow().isoformat() + 'Z'
    print(dict_body['event_id'])
    print(dict_body['timestamp'])

    bytes_io_body = compress_gzip(dict_body)
    
    # print(bytes_io_body.getvalue())
    # print(row_proxy.headers)

    try:
        # bytes_io_body.getvalue() is for reading the bytes
        response = http.request(
            "POST", str(SENTRY_API_STORE_ONPREMISE), body=bytes_io_body.getvalue(), headers=row_proxy.headers 
        )
    except Exception as err:
        print('LOCAL EXCEPTION', err)

    return("FINISH")
    return 'loaded and forwarded to Sentry'

##########################  TESTING  ###############################

# STEP1 - TESTING w/ database. send body {"foo": "bar"} from Postman
@app.route('/save-event', methods=['POST'])
def event_bytea_post():

    request_headers = {}
    for key in ['Host','Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent']:
        request_headers[key] = request.headers.get(key)
    print('request_headers', request_headers)

    insert_query = """ INSERT INTO events (type, name, data, headers) VALUES (%s,%s,%s,%s)"""
    record = ('python', 'example', request.data, json.dumps(request_headers))

    with db.connect() as conn:
        conn.execute(insert_query, record)
        conn.close()
    return 'successfull bytea'

# STEP 2 - TESTING w/ database. loads that event's bytes+headers from database
@app.route('/load-event', defaults={'pk':0}, methods=['GET'])
@app.route('/load-event/<pk>', methods=['GET'])
def event_bytea_get():

    if pk==0:
        query = "SELECT * FROM events ORDER BY pk DESC LIMIT 1;"
    else:      # bytes_io_body.getvalue() is for reading the bytes
        query = "SELECT * FROM events WHERE pk={};".format(pk)

    with db.connect() as conn:
        results = conn.execute(query).fetchall()
        conn.close()
        row_proxy = results[0]

        return { "data": decompress_gzip(row_proxy.data), "headers": row_proxy.headers }
        # return { "data": row_proxy.data.decode("utf-8"), "headers": row_proxy.headers }
