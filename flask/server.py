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
# SENTRY_API_STORE = "https://18562a9e8e3943088b1ca3cedf222e21@sentry.io/api/1435220/store/"
# http://759bf0ad07984bb3941e677b35a13d2c@localhost:9000/2

# TODO see what sentry-sdk sends as outbound url
# TODO does this match the numbers in dsn on onpremise sentry? check UI
SENTRY_API_STORE_ONPREMISE ="http://759bf0ad07984bb3941e677b35a13d2c@localhost:9000/api/2/store"
# SENTRY_API_STORE_ONPREMISE ="http://localhost:9000/api/2/store"

app = Flask(__name__)
CORS(app)

# Receives payload from sentry-sdk in app.py
@app.route('/api/2/store/', methods=['POST'])
def event():
    print('\n type request.data', type(request.data))
    print('\n content_encoding', request.content_encoding)
    print('\n content_type', request.content_type)

    headers = request.headers
    # for key in headers:
    #     print("key", key)
    requests_headers = {
        # 'Host': headers.get('Host'),
        'Host': headers.get('Host'),
        # 'Cookie': '_xsrf=2|978a1d70|6e2dc68178acf21053e8d6bd21b7d063|1583120785; username-localhost-8889="2|1:0|10:1583896358|23:username-localhost-8889|44:NDhmMGU4YjEyMjNjNDhiZGFlMjgwY2Y5ZTQ2ZjEzZmI=|5891415757f2c0073dc30c5cadeaa886a772bd22c3c6afba31e0ef28ace2d089"; sc=p9v2tVJRLp6n5EiS19Qci6EXdtsBFAsaBfolfuTwGHx1IKwSCmF4YPoF1OCEvb22; username-localhost-8888="2|1:0|10:1584890480|23:username-localhost-8888|44:N2JmZDI0MTczYzkzNDQ0ZDhjZGYzNTY2M2E1MTE1OWI=|5386d1c09598742716d6e4e505c85b02eb7b3180c160093127ffd7ca33012f0b"; io=Yl3cJ51IKH6ILCp9AAAI; sentrysid=".eJxNjssKgkAUhs0iKIqgR2jVSpTxksuC3qAhdzKeM-SgOTWXLougR8_KhbvD9___x3m7r-tgTec5s6bMreYqF5gNHMcJ6ISBETcu1Skbt0Dzxqjn1aUjbVFmsxYdHuaI5gK3XaCyaQsM1wakrAT_Te5SVRzpsmcvGFS8Qbr66zxrRK29b-7tz0zU2_badZ2h1rL7ZdFTlEyXdI3ICBIf04IACzEIgfgbkiIibPw4hihK0iSF2HofbvNIkg:1jG9pT:JXPuM56EKosEyJXHhlwjByyUA-E"',
        'Accept-Encoding': headers.get('Accept-Encoding'),

        'Content-Length': headers.get('Content-Length'),

        'Content-Encoding': headers.get('Content-Encoding'),
        'Content-Type': 'application/json',
        # 'X-Sentry-Auth': headers.get('X-Sentry-Auth'),
        'X-Sentry-Auth': 'sentry_key=18562a9e8e3943088b1ca3cedf222e21, sentry_version=7, sentry_client=sentry.python/0.14.2',
        'User-Agent': headers.get('User-Agent'),

        # 'Referer': 'localhost:9000'
    }
    print("request.headers", request.headers)
    print("requests_headers", requests_headers)

    # TODO Save data and headers to DB
    data = decompress_gzip(request.data) # keys: exception, server_name, tags, event_id, timestamp, extra, modules, contexts, platform, breadcrumbs, level, sdk
    for key in json.loads(data):
        print("key", key)
    # JSON
    # print('dump', json.dumps(data, indent=1).replace("\\",""))
    # print ('request.json', request.json) # No, it's bytes

    try:
        # request.data.decode('utf-8') request.get_data().decode('utf-8')  request.get_data('as_text'))
        # http.request('POST', SENTRY_API_STORE_ONPREMISE, fields=json.dumps(data, indent=2).replace("\\",""), headers=requests_headers)

        # TODO  
        body = io.BytesIO()
        with gzip.GzipFile(fileobj=body, mode="w") as f:
            f.write(json.dumps(data, allow_nan=False).encode("utf-8"))

        print('type(request.headers)', type(request.headers)) #flask type...BAD
        print('type(requests_headers)', type(requests_headers)) #dict
        print('type(body.getvalue()', type(body.getvalue())) # why does this print as 'str'?

        #ca_cert stuff?
        response = http.request(
            "POST", str(SENTRY_API_STORE_ONPREMISE), body=body.getvalue(), headers=requests_headers 
        )

        print("\nresponse.data", response.data)
        print("\nr.status", response.status)


    except Exception as err:
        print('\nLOCAL EXCEPTION')
        print(err)
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

# DSN = "http://759bf0ad07984bb3941e677b35a13d2c@localhost:9000/2"
# sentry_sdk.init(dsn=DSN)


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