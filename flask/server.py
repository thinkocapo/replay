import os
from flask import Flask, request, json, abort
from flask_cors import CORS
import requests
import sentry_sdk
from sentry_sdk.integrations.flask import FlaskIntegration
import gzip
import io
from six import BytesIO
from gzip import GzipFile
import urllib3
http = urllib3.PoolManager()

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

# def get_connection():
#     with sentry_sdk.start_span(op="psycopg2.connect"):
#         connection = psycopg2.connect(
#             host = "",
#             database = "",
#             user = "",
#             password = "")
#     return connection 

# Intercepts the payload sent by sentry_sdk in app.py, and then sends it to a Sentry instance
@app.route('/api/2/store/', methods=['POST'])
def event():
    # print(request.headers)
    headers = request.headers
    requests_headers = {
        'Host': headers.get('Host'),
        'Accept-Encoding': headers.get('Accept-Encoding'), # 'identity' ? 'gzip' '*'
        # 'Content-Length': headers.get('Content-Length'), # ?
        'Content-Encoding': headers.get('Content-Encoding'),
        'Content-Type': headers.get('Content-Type'),
        'User-Agent': headers.get('User-Agent')
    }
    # print('requests_headers', requests_headers)
    print('type(request.data)', type(request.data)) # bytes
    # print('missing padding?', len(request.data) % 4)

    data = decompress_gzip(request.data) 
    print('type(data)', type(data)) # string / JSON

    # TODO - save to postgres 
    # try:
    #     db = psycopg2.connect(*args)  # DB-API 2.0
    #     c = db.cursor()
    #     c.execute('''INSERT INTO test VALUES (%s, %s, %s)''', (bin_data, 1337, 'foo'))
    #     print('try')
    # except Exception as err:
    #     print('LOCAL EXCEPTION', err)

    try:
        body = io.BytesIO()
        with gzip.GzipFile(fileobj=body, mode="w") as f:
            f.write(json.dumps(data, allow_nan=False).encode("utf-8"))
            f.close()

        print('type(body)', type(body))
        body = body.getvalue()
        print('type(body.getvalue())', type(body))

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
        fp = BytesIO(encoded_data) # + "==" or "="
        try:
            f = GzipFile(fileobj=fp)
            return f.read().decode("utf-8")
        finally:
            f.close()
    except Exception as e:
        raise e

@app.route('/test', methods=['GET'])
def test():
    return 'Success'