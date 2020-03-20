import os
from flask import Flask, request, json, abort
from flask_cors import CORS
import requests
import urllib3
http = urllib3.PoolManager()
import sentry_sdk
from sentry_sdk.integrations.flask import FlaskIntegration
import gzip
import io
from six import BytesIO
from gzip import GzipFile

# Unmodified DSN for  'python-eat' Project in testorg-az
# https://18562a9e8e3943088b1ca3cedf222e21@sentry.io/1435220

# Sentry Server endpoint for receiving events for above DSN
SENTRY_API_STORE = "https://18562a9e8e3943088b1ca3cedf222e21@sentry.io/api/1435220/store/"
SENTRY_API_STORE_ONPREMISE ="http://759bf0ad07984bb3941e677b35a13d2c@localhost:9000/api/2/store"
DSN = "http://759bf0ad07984bb3941e677b35a13d2c@localhost:9000/2"
sentry_sdk.init(dsn=DSN)

app = Flask(__name__)
CORS(app)

@app.route('/api/42/store/', methods=['POST'])
def event():
    print('\n type request.data', type(request.data))
    print('\n content_encoding', request.content_encoding)
    print('\n content_type', request.content_type)

    headers = request.headers
    for key in headers:
        print("key", key)

    requests_headers = {
        'Host': request.headers.get('Host'),
        'Accept-Encoding': request.headers.get('Accept-Encoding'),
        'Content-Length': request.headers.get('Content-Length'),
        'Content-Encoding': request.headers.get('Content-Encoding'),
        'Content-Type': request.headers.get('application/json'),
        'X-Sentry-Auth': request.headers.get('X-Sentry-Auth'),
        'User-Agent': request.headers.get('User-Agent')
    }

    # TODO Save data and headers to DB
    data = decompress_gzip(request.data)
    # keys: exception, server_name, tags, event_id, timestamp, extra, modules, contexts, platform, breadcrumbs, level, sdk
    # for key in json.loads(data):
        # print("key", key)

    # JSON
    # print('dump', json.dumps(data, indent=1).replace("\\",""))
    # print ('request.json', request.json) # No, it's bytes

    # Send to a Sentry instance
    try:
        print('00000')

        # request.data.decode('utf-8') request.get_data().decode('utf-8')  request.get_data('as_text'))

        # r = requests.post(
        #     SENTRY_API_STORE_ONPREMISE, 
        #     json=json.dumps(data, indent=2).replace("\\",""),
        #     # data=request.data,
        #     headers=requests_headers
        # )

        # http.request('POST', SENTRY_API_STORE_ONPREMISE, fields=json.dumps(data, indent=2).replace("\\",""), headers=requests_headers)

        # TODO  
        # body = io.BytesIO()
        # with gzip.GzipFile(fileobj=body, mode="w") as f:
        #     f.write(json.dumps(json.dumps(data, indent=1).replace("\\",""), allow_nan=False).encode("utf-8"))
        # http.request('POST', SENTRY_API_STORE_ONPREMISE, fields=body.getvalue(), headers={"Content-Type": "application/json", "Content-Encoding": "gzip"})
        
        response = self._pool.request(
            "POST", str(self._auth.store_api_url), body=body, headers=headers
        )
    
    
        print('11111')
    except Exception as err:
        print('\nEXCEPTION')
        print(err)
        # sentry_sdk.capture_exception(err) # <-- works but is not the intent for this repo
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


# body = io.BytesIO()
# with gzip.GzipFile(fileobj=body, mode="w") as f:
#     f.write(json.dumps(event, allow_nan=False).encode("utf-8"))

# assert self.parsed_dsn is not None
# logger.debug(
#     "Sending event, type:%s level:%s event_id:%s project:%s host:%s"
#     % (
#         event.get("type") or "null",
#         event.get("level") or "null",
#         event.get("event_id") or "null",
#         self.parsed_dsn.project_id,
#         self.parsed_dsn.host,
#     )
# )
# self._send_request(
#     body.getvalue(),
#     headers={"Content-Type": "application/json", "Content-Encoding": "gzip"},
# )
# return None

# TODO
# def _send_request(
#     self,
#     body,  # type: bytes
#     headers,  # type: Dict[str, str]
# ):
#     # type: (...) -> None
#     headers.update(
#         {
#             "User-Agent": str(self._auth.client),
#             "X-Sentry-Auth": str(self._auth.to_header()),
#         }
#     )
#     response = self._pool.request(
#         "POST", str(self._auth.store_api_url), body=body, headers=headers
#     )