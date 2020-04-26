import os
import datetime
from dotenv import load_dotenv
import gzip
from gzip import GzipFile
import io
import json
# import sentry_sdk
# from sentry_sdk.integrations.flask import FlaskIntegration
from services import compress_gzip, decompress_gzip
from six import BytesIO
import sqlite3
import sys
import urllib3
import uuid
load_dotenv()
http = urllib3.PoolManager()

# DATABASE
SQLITE = os.getenv('SQLITE')
database = SQLITE or os.getcwd() + "/sqlite.db"
print('> database', database)

DSN = os.getenv('DSN')
KEY = DSN.split('@')[0][7:]
SENTRY ="http://localhost:9000/api/2/store/?sentry_key={}&sentry_version=7".format(KEY)
print('> Sentry', SENTRY)

"""the final http request to Sentry here hangs sometimes. haven't figured out why. event-to-sentry.go works more consistently"""

with sqlite3.connect(database) as db:

    cursor = db.cursor()
    _id = sys.argv[1] if len(sys.argv) > 1 else None
    if _id==None:
        cursor.execute("SELECT * FROM events ORDER BY id DESC LIMIT 1;")
    else:
        cursor.execute("SELECT * FROM events WHERE id=?", [_id])
    rows = cursor.fetchall()
    row = rows[0]
    row = list(row) # ?
    body_bytes_buffer = row[3]  
    request_headers = json.loads(row[4])

    # update event_id/timestamp so Sentry will accept the event again
    json_body = decompress_gzip(body_bytes_buffer)
    dict_body = json.loads(json_body)
    print('dict_body value:', dict_body['exception']['values'][0]['value'])
    dict_body['event_id'] = uuid.uuid4().hex
    dict_body['timestamp'] = datetime.datetime.utcnow().isoformat() + 'Z'
    print('> event_id', dict_body['event_id'])
    print('> timestamp', dict_body['timestamp'])
    bytes_io_body = compress_gzip(dict_body)
        
try:
    # bytes_io_body.getvalue() is for reading the bytes
    response = http.request(
        "POST", str(SENTRY), body=bytes_io_body.getvalue(), headers=request_headers
    )
    response.close()
except Exception as err:
    print('LOCAL EXCEPTION', err)