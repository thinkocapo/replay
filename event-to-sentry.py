import os
import datetime
from dotenv import load_dotenv
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

http = urllib3.PoolManager()

SENTRY ="http://localhost:9000/api/2/store/?sentry_key=09aa0d909232457a8a6dfff118bac658&sentry_version=7"

# DATABASE
SQLITE = os.getenv('SQLITE')
database = SQLITE or os.getcwd() + "/sqlite.db"
print('> database', database)


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

with sqlite3.connect(database) as conn:

    cur = conn.cursor()
    cur.execute("SELECT * FROM events ORDER BY id DESC LIMIT 1;")
    rows = cur.fetchall()
    row = rows[0]
    row = list(row) # ?
    body_bytes_buffer = row[3] # not row_proxy.data, because sqlite returns tuple (not row_proxy)
    request_headers = json.loads(row[4])
    # print('\n type(body_bytes_buffer)', type(body_bytes_buffer))

    # call it 'bytes_buffer_body'
    # update event_id/timestamp so Sentry will accept the event again
    json_body = decompress_gzip(body_bytes_buffer)
    dict_body = json.loads(json_body)
    dict_body['event_id'] = uuid.uuid4().hex
    dict_body['timestamp'] = datetime.datetime.utcnow().isoformat() + 'Z'
    print('> event_id', dict_body['event_id'])
    print('> timestamp', dict_body['timestamp'])
    print('DICT', type(dict_body))
    bytes_io_body = compress_gzip(dict_body)
        
try:
    print('type(bytes_io_body)', type(bytes_io_body))
    print('type(bytes_io_body.getvalue())', type(bytes_io_body.getvalue()))
    
    # print('VALUE', bytes_io_body.getvalue())
    
    # bytes_io_body.getvalue() is for reading the bytes
    # WORKS on python3
    response = http.request(
        # "POST", str(SENTRY), body=bytearray(bytes_io_body.getvalue()), headers=request_headers
        "POST", str(SENTRY), body=bytes_io_body.getvalue(), headers=request_headers
    )
    response.close()
except Exception as err:
    print('LOCAL EXCEPTION', err)