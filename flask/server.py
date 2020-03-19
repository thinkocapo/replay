import os
from flask import Flask, request, json, abort
from flask_cors import CORS
import requests
import sentry_sdk
from sentry_sdk.integrations.flask import FlaskIntegration
import gzip
from six import BytesIO
from gzip import GzipFile

# Unmodified DSN for  'python-eat' Project in testorg-az
# https://18562a9e8e3943088b1ca3cedf222e21@sentry.io/1435220

# Sentry Server endpoint for receiving events for above DSN
API_STORE = "https://18562a9e8e3943088b1ca3cedf222e21@sentry.io/api/1435220/store/"

app = Flask(__name__)
CORS(app)

@app.route('/api/42/store/', methods=['POST'])
def event():
    print('Headers', request.headers)

    # Must decompress gzp in order to print the data
    # Prints the keys: exception, server_name, tags, event_id, timestamp, extra, modules, contexts, platform, breadcrumbs, level, sdk
    data = decompress_gzip(request.data)
    for key in json.loads(data):
        print("key", key)

    # TODO Save request.data and request.headers to DB
    # TODO Send it to Sentry.io. add the right headers and use urllib3
    try:
        r = requests.post(API_STORE, data=request.data)
    except Exception as err:
        print(err)
        return 'failed'
    return 'Success - handled'

@app.route('/test', methods=['GET'])
def test():
    try:
        return 'Success'
    except Exception as err:
        print(err)
        return 'Failure'

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
