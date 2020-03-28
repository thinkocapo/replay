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
    insert_query = """ INSERT INTO events (type, name) VALUES (%s,%s)"""
    record = ('python', 'example')
    with db.connect() as conn:
        conn.execute(
            "INSERT INTO events (type,name) VALUES ('type2', 'name2')"
        )
        conn.close()
        print("inserted")
        # rows = []
        # for row in results:
        #     rows.append(dict(row))
        # return json.dumps(rows)

@app.route('/event-bytea', methods=['POST'])
def event_bytea():
    binary = request

    insert_query = """ INSERT INTO events (type, name, data) VALUES (%s,%s,%s)"""
    record = ('python', 'example', binary)
    
    with db.connect() as conn:
        conn.execute(insert_query, record))
        conn.close()

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