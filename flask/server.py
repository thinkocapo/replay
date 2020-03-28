import os
from flask import Flask, request, json, abort
from flask_cors import CORS
import gzip
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

''' NOTES
Got error w/ 403-csrf.html until I put X-Sentry-Auth in URL rather than headers, which then gave error on the onprem Internal project
But (my interpretation of) getsentry/sentry-python shows it being set in the request's headers, not URL.
    # 'X-Sentry-Auth': headers.get('X-Sentry-Auth'),
    # 'X-Sentry-Auth': 'Sentry sentry_key=759bf0ad07984bb3941e677b35a13d2c, sentry_version=7, sentry_client=sentry.python/0.14.2',
'''

# SENTRY_API_STORE_ONPREMISE ="http://localhost:9000/api/2/store"
SENTRY_API_STORE_ONPREMISE ="http://localhost:9000/api/2/store/?sentry_key=759bf0ad07984bb3941e677b35a13d2c&sentry_version=7"

app = Flask(__name__)
CORS(app)

HOST='localhost'
DATABASE='postgres'
USERNAME='admin'
PASSWORD='admin'
db = create_engine('postgresql://' + USERNAME + ':' + PASSWORD + '@' + HOST + ':5432/' + DATABASE)

# Intercepts the payload sent by sentry_sdk in app.py, and then sends it to a Sentry instance
@app.route('/api/2/store/', methods=['POST'])
def api_store():
    print('type(request)', type(request))
    print('type(request.headers)', type(request.headers))
    print('type(request.data)', type(request.data))

    headers = request.headers
    requests_headers = {
        'Host': headers.get('Host'),
        'Accept-Encoding': headers.get('Accept-Encoding'),
        'Content-Length': headers.get('Content-Length'),
        'Content-Encoding': headers.get('Content-Encoding'),
        'Content-Type': headers.get('Content-Type'),
        'User-Agent': headers.get('User-Agent')
    }

    data = decompress_gzip(request.data)
    print('type(data)', type(data))
    print('data', data)

    try:
        body = io.BytesIO()
        with gzip.GzipFile(fileobj=body, mode="w") as f:
            f.write(json.dumps(data, allow_nan=False).encode("utf-8"))

        # TODO body=body.getvalue() errors in the onprem Sentry as "b'{"error":"Bad data decoding request (TypeError, Incorrect padding)"}'"
        response = http.request(
            "POST", str(SENTRY_API_STORE_ONPREMISE), body=request.data, headers=requests_headers 
        )

        print("%s RESPONSE and event_id %s" % (response.status, response.data))
        return 'success'
    except Exception as err:
        print('LOCAL EXCEPTION', err)


def decompress_gzip(encoded_data):
    try:
        fp = BytesIO(encoded_data)
        try:
            f = GzipFile(fileobj=fp)
            return f.read().decode("utf-8")
        finally:
            f.close()
    except Exception as e:
        raise e

def get_connection():
    with sentry_sdk.start_span(op="psycopg2.connect"):
        connection = psycopg2.connect(
            host=HOST,
            database=DATABASE,
            user=USERNAME,
            password=PASSWORD)
    return connection 

@app.route('/event-bytea', methods=['GET'])
def event_bytea_get():
    print('/event GET')

    # set typecasting because psycopg2 will return a <MemoryView> for bytea instead of the bytes
    def bytea2bytes(value, cur):
        m = psycopg2.BINARY(value, cur)
        if m is not None:
            return m.tobytes()
    BYTEA2BYTES = psycopg2.extensions.new_type(
        psycopg2.BINARY.values, 'BYTEA2BYTES', bytea2bytes)
    psycopg2.extensions.register_type(BYTEA2BYTES)

    with db.connect() as conn:
        results = conn.execute(
            "SELECT * FROM events WHERE pk=11"
        ).fetchall()
        conn.close()
        print('results[0]', results[0])

        row_proxy = results[0]
        
        print('type(row_proxy)', type(row_proxy))
        print('row_proxy', row_proxy)
        keys = row_proxy.keys()
 
        for key in keys:
            print("key", key)

        print('row_proxy.type', row_proxy.type)
        print('row_proxy.data', row_proxy.data)
        print('type(row_proxy.data)', type(row_proxy.data)) #'bytes' if you use the typecasting. 'MemoryView' if you don't use typecasting
        print('row_proxy.data', row_proxy.data)

        return row_proxy.data
        # strings = decompress_gzip(row_proxy.data)
        # print('strings', strings)
        
        # rows = []
        # for row in results:
        #     rows.append(dict(row))
        # return json.dumps(rows)

@app.route('/event-bytea', methods=['POST'])
def event_bytea_post():
    # TODO different from request and request.headre but try it
    print('/event-bytea POST')
    print('type(request.data)', type(request.data)) # bytes
    print('request.data', request.data)

    # fp = BytesIO(request.data)
    # print('type(fp)', type(fp))

    insert_query = """ INSERT INTO events (type, name, data) VALUES (%s,%s,%s)"""
    record = ('python', 'example', request.data)
    # record = ('python', 'example', fp)

    with db.connect() as conn:
        conn.execute(insert_query, record)
        conn.close()
    return 'successfull bytea'

@app.route('/events', methods=['GET'])
def events():
    print('/event GET')

    with db.connect() as conn:
        results = conn.execute(
            "SELECT * FROM events"
        ).fetchall()
        conn.close()
        
        rows = []
        for row in results:
            rows.append(dict(row))
        return json.dumps(rows)

@app.route('/event', methods=['POST'])
def event():
    print('/event POST')
    record = ('python', 'example')
    insert_query = """ INSERT INTO events (type, name) VALUES (%s,%s)"""
    with db.connect() as conn:
        conn.execute(
            "INSERT INTO events (type,name) VALUES ('type4', 'name4')"
        )
        conn.close()
        print("inserted")
    return 'successful'

@app.route('/test', methods=['GET'])
def test():
    return 'Success'

# connection = get_connection()
# cursor = connection.cursor(cursor_factory = psycopg2.extras.DictCursor)
# try:
#     cursor.execute(insert_query, (name, tool_type, randomString(10), image, random.randint(10,50)))
#     connection.commit()
# except:
#     raise "Row insert failed\n"
#     return 'fail'
# cursor.close()
# connection.close()