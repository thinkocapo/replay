import os
import datetime
from dotenv import load_dotenv
from flask import Flask, request, json, abort
from flask_cors import CORS
import json
from services import compress_gzip, decompress_gzip, get_event_type
import sqlite3
import string # ?
import urllib3
import uuid
load_dotenv()
http = urllib3.PoolManager()

app = Flask(__name__)

app.run(threaded=True)
CORS(app)

print("""
                               Welcome To The
  _   _   _   _   ____    _____   ____    _____      _      _  __  _____   ____  
 | | | | | \ | | |  _ \  | ____| |  _ \  |_   _|    / \    | |/ / | ____| |  _ \ 
 | | | | |  \| | | | | | |  _|   | |_) |   | |     / _ \   | ' /  |  _|   | |_) |
 | |_| | | |\  | | |_| | | |___  |  _ <    | |    / ___ \  | . \  | |___  |  _ < 
  \___/  |_| \_| |____/  |_____| |_| \_\   |_|   /_/   \_\ |_|\_\ |_____| |_| \_\
                                                                                 
""")

SENTRY=''

JSON = os.getenv('JSON') # or os.getcwd() + "/sqlite.db"
print("> json database is:", JSON)

JSON_TRANSACTIONS = "transactions.json"

# OG
# MODIFIED_DSN_FORWARD - Intercepts the payload sent by sentry_sdk in event.py, and then sends it to a Sentry instance
@app.route('/api/2/store/', methods=['POST'])
def forward():
    print('> FORWARD')

    # TODO exception.platform may have been available, as well as exception.sdk https://github.com/thinkocapo/undertaker/issues/48
    def make(headers):
        request_headers = {}
        user_agent = request.headers.get('User-Agent').lower()
        if 'python' in user_agent:
            print('> Python error')
            for key in ['Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent']:
                request_headers[key] = request.headers.get(key)
                # SENTRY = sentryUrl(os.getenv('DSN_PYTHON'))
                SENTRY = sentryUrl(os.getenv('DSN_PYTHONTEST'))
        if 'mozilla' in user_agent or 'chrome' in user_agent or 'safari' in user_agent:
            print('> Javascript error')
            for key in ['Accept-Encoding','Content-Length','Content-Type','User-Agent']:
                request_headers[key] = request.headers.get(key)
                # SENTRY = sentryUrl(os.getenv('DSN_REACT'))
                SENTRY = sentryUrl(os.getenv('DSN_REACT_SAAS'))
        return request_headers, SENTRY

    request_headers, SENTRY = make(request.headers)
    print('> SENTRY url for store endpoint', SENTRY)

    try:
        print('> type(request.data)', type(request.data))
        print('> type(request_headers)', type(request_headers))

        response = http.request(
            "POST", str(SENTRY), body=request.data, headers=request_headers 
        )

        print('> nothing saved to sqlite database')
        return 'success'
    except Exception as err:
        print('LOCAL EXCEPTION', err)

# MODIFIED_DSN_FORWARD 
@app.route('/api/2/envelope/', methods=['POST'])
def forward_envelope():
    print('> /api/2/envelope/ FORWARD')
    
    def make(headers):
        request_headers = {}
        user_agent = request.headers.get('User-Agent').lower()
        if 'python' in user_agent:
            print('> Python envelope')
            for key in ['Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent']:
                request_headers[key] = request.headers.get(key)
                SENTRY = sentryUrlEnvelope(os.getenv('DSN_PYTHON_SAAS'))
        if 'mozilla' in user_agent or 'chrome' in user_agent or 'safari' in user_agent:
            print('> Javascript envelope')
            for key in ['Accept-Encoding','Content-Length','Content-Type','User-Agent']:
                request_headers[key] = request.headers.get(key)
                SENTRY = sentryUrlEnvelope(os.getenv('DSN_JAVASCRIPT_SAAS'))
        return request_headers, SENTRY

    request_headers, SENTRY = make(request.headers)
    print('> SENTRY url for store endpoint', SENTRY)

    try:
        print('> type(request.data)', type(request.data))
        print('> type(request_headers)', type(request_headers))

        response = http.request(
            "POST", str(SENTRY), body=request.data, headers=request_headers 
        )

        print('> nothing saved to json file')
        return 'success1'
    except Exception as err:
        print('LOCAL EXCEPTION', err)

# OG
# MODIFIED_DSN_SAVE - Intercepts event from sentry sdk and saves them to json file. No forward of event to your Sentry instance.
@app.route('/api/3/store/', methods=['POST'])
def save():
    print('> SAVING')

    event_platform = ''
    event_type = ''
    request_headers = {}
    user_agent = request.headers.get('User-Agent').lower()
    body = ''

    if 'python' in user_agent:
        event_platform = 'python'
        event_type = "error"
        print('> PYTHON', event_type)
        for key in ['Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent', 'X-Sentry-Auth']:
            request_headers[key] = request.headers.get(key)
        body = decompress_gzip(request.data)
    if 'mozilla' in user_agent or 'chrome' in user_agent or 'safari' in user_agent:
        event_platform = 'javascript'
        event_type = "error"
        print('> JAVASCRIPT ', event_type)
        for key in ['Accept-Encoding','Content-Length','Content-Type','User-Agent']:
            request_headers[key] = request.headers.get(key)
        body = request.data

    # TODO store the bytes again...
    # body = json.loads(body)

    print("TYPE", type(request.data))
    print("REQUEST", type(request.data.decode('utf8')))
    # print(bytes(body))

    o = {
        'platform': event_platform,
        'kind': event_type,
        'headers': request_headers, # json.dumps(request_headers),
        'body': str(request.data)
        # 'body': request.data
        # 'body': request.data.decode('utf8') # writes as "body": "{\"exception\":{\"values\":[{\"
        # 'body': json.dumps(request.data)
        # 'body': request.data.decode('ascii') # writes as "body": "{\"exception\":{\"values\":[{\"
        # 'body': bytes(str(body), "utf8")
        # 'body': bytes(json.dumps(body), "utf8"),
        # 'body': bytes(body)
        # 'body': json.loads(body)
    }

    try:
        with open(JSON) as file:
            current_data = json.load(file)
        with open(JSON, 'w') as file:
            current_data.append(o)
            json.dump(current_data, file)

    except Exception as exception:
        print("LOCAL EXCEPTION", exception)
    return "success"

# MODIFIED_DSN_SAVE
@app.route('/api/3/envelope/', methods=['POST'])
def save_envelope():
    print('> SAVING')

    event_platform = ''
    event_type = "transaction"
    request_headers = {}
    user_agent = request.headers.get('User-Agent').lower()
    body = ''

    if 'python' in user_agent:
        event_platform = 'python'
        print('> PYTHON', event_type)
        for key in ['Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent', 'X-Sentry-Auth']:
            request_headers[key] = request.headers.get(key)
        body = decompress_gzip(request.data)
        # body = body.split('\n')
    if 'mozilla' in user_agent or 'chrome' in user_agent or 'safari' in user_agent:
        event_platform = 'javascript'
        print('> JAVASCRIPT ', event_type)
        for key in ['Accept-Encoding','Content-Length','Content-Type','User-Agent']:
            request_headers[key] = request.headers.get(key)
        #body = request.data.decode("utf-8")
        # print('BODY BEFORE', body) # still no slashes
        # body = body.replace("\\", "") # not needed, as slashes are addd when getting saved
        #body = body.split('\n') # not needed since storing bytes (of the envelope string). .split turns it into a List

    print("\n> TYPE OF BODY: ", type(body))
    
    # for item in body:
        # print(type(item))
        # item = json.loads(item)
        # print(type(item))

    # print(json.dumps(body))
    
    # TODO store as slice of bytes?
    o = {
        'platform': event_platform,
        'kind': event_type,
        'headers': request_headers,
        'body': body
        # 'body': json.loads(body)
        # 'body': json.dumps(body) # adds too many slashes
    }

    try:
        with open(JSON_TRANSACTIONS) as file:
            current_data = json.load(file)

        with open(JSON_TRANSACTIONS, 'r+') as file:
            # current_data = json.load(file)

            current_data.append(o)
            json.dump(current_data, file)

    except Exception as exception:
        print("LOCAL EXCEPTION", exception)
    return "success"




# FORWARD TRANSACTION?
@app.route('/api/6/store/', methods=['POST'])
def forward_store():
    print('> /api/6/store/ FORWARD')
    # print('> request.headers', request.headers)

    def make(headers):
        SENTRY = sentryUrl(os.getenv('DSN_ANDROID'))
        request_headers = {}
        for key in ['X-Sentry-Auth', 'Content-Length','User-Agent','X-Forwarded-Proto','Host','Accept','X-Forwarded-For', 'Content-Type','Accept-Encoding']:
                request_headers[key] = request.headers.get(key)
        return request_headers, SENTRY
        
    request_headers, SENTRY = make(request.headers)
    print('> SENTRY url for store endpoint', SENTRY)

    try:
        print('> type(request.data)', type(request.data))
        print('> type(request_headers)', type(request_headers))
        # print('> request_headers', request_headers)

        response = http.request(
            "POST", str(SENTRY), body=request.data, headers=request_headers 
        )

        # print('> nothing saved to sqlite database')
        return 'success'
    except Exception as err:
        print('LOCAL EXCEPTION', err)

    return 'good'

# SAVE ANDROID (DELETE?)
@app.route('/api/8/envelope/', methods=['POST'])
def save_mobile_envelope():

    print('\n> /api/5/envelope ')
    print('> type(request.data)', type(request.data)) # string
    print('> type(request.headers)', type(request.headers))

    event_platform = ''
    event_type = ''
    request_headers = {}
    user_agent = request.headers.get('User-Agent').lower()
    body = ''

    event_platform = 'android'
    event_type = get_event_type(request.data, "android")
    print('\n> event_type', event_type)
    
    for key in ['X-Sentry-Auth', 'Content-Length','User-Agent','Connection','Content-Encoding','X-Forwarded-Proto','Host','Accept','X-Forwarded-For', 'Content-Type', 'Accept-Encoding']:
        request_headers[key] = request.headers.get(key)
    
    print("type", type(decompress_gzip(request.data))) # <type 'unicode'>
    print(decompress_gzip(request.data))

    # decompressed = decompress_gzip(request.data)
    # converted = decompressed.encode("utf-8")
    # print("type of converted", type(converted)) # <type 'string'>
    # print(converted)

    body = decompress_gzip(request.data)

    insert_query = ''' INSERT INTO events(platform,type,body,headers)
              VALUES(?,?,?,?) '''
    record = (event_platform, event_type, body, json.dumps(request_headers))
    try:
        with sqlite3.connect(database) as db:
            cursor = db.cursor()
            cursor.execute(insert_query, record)
            print('> SQLITE ID', cursor.lastrowid)
            cursor.close()
            return str(cursor.lastrowid)
    except Exception as err:
        print("LOCAL EXCEPTION", err)

    print('> SAVING /api/5/envelope END')

    return 'SUCCESS'

# FORWARD ANDROID
# @app.route('/api/9/envelope/', methods=['POST'])
# def forward_envelope():
#     print('> /api/6/envelope/ FORWARD')

#     def make(headers):
#         print('0000000')
#         SENTRY = sentryUrl(os.getenv('DSN_ANDROID'))
#         request_headers = {}
#         for key in ['X-Sentry-Auth', 'Content-Length','User-Agent','Connection','Content-Encoding','X-Forwarded-Proto','Host','Accept','X-Forwarded-For', 'Content-Type', 'Accept-Encoding']:
#             print('11111', key)            
#             request_headers[key] = request.headers.get(key)
#         return request_headers, SENTRY

#     request_headers, SENTRY = make(request.headers)
#     print('> SENTRY url for store endpoint', SENTRY)
#     try:
#         print('> type(request.data)', type(request.data))
#         print('> type(request_headers)', type(request_headers))

#         response = http.request(
#             "POST", str(SENTRY), body=request.data, headers=request_headers 
#         )

#         print('> nothing saved to sqlite database')
#         return 'success'
#     except Exception as err:
#         print('LOCAL EXCEPTION', err)

#     return 'good'

# MODIFIED_DSN_SAVE MOBILE - Intercepts event from sentry sdk and saves them to Sqlite DB. No forward of event to your Sentry instance.
@app.route('/api/8/store/', methods=['POST'])
def save_mobile():
    print('> /api/5/store')

    print('> type(request.data)', type(request.data)) # STRING
    print('> type(request_headers)', type(request.headers))

    event_platform = ''
    event_type = ''
    request_headers = {}
    user_agent = request.headers.get('User-Agent').lower()
    body = ''

    event_platform = 'android'
    event_type = get_event_type(request.data, "android")
    print('> event_type', event_type)
    

    # ATODO are these different if it's a session?
    for key in ['X-Sentry-Auth', 'Content-Length','User-Agent','Connection','Content-Encoding','X-Forwarded-Proto','Host','Accept','X-Forwarded-For']:
        request_headers[key] = request.headers.get(key)
    # print(json.dumps(request_headers,indent=2))
    # print(json.dumps(json.loads(decompress_gzip(request.data)),indent=2))
    # body = decompress_gzip(request.data) # 'error: not a gzipped file' in decompress_gzip

    # ATODO verify always right
    body = request.data

    insert_query = ''' INSERT INTO events(platform,type,body,headers)
              VALUES(?,?,?,?) '''
    record = (event_platform, event_type, body, json.dumps(request_headers))
    try:
        with sqlite3.connect(database) as db:
            cursor = db.cursor()
            cursor.execute(insert_query, record)
            print('> SQLITE ID', cursor.lastrowid)
            cursor.close()
            return str(cursor.lastrowid)
    except Exception as err:
        print("LOCAL EXCEPTION", err)

    print('> SAVING /api/5/store END')
    return 'SUCCESS'

###############################################################################


# OG
# MODIFIED_DSN_SAVE_AND_FORWARD - this has been out of date since proxy.py started supporting Transactions in /api/2/store and /api/3/store endpoints
@app.route('/api/4/store/', methods=['POST'])
def save_and_forward():

    # print('> type(request.data)', type(request.data))
    # print('> type(request_headers)', type(request.headers))
    # for header in request.headers.to_wsgi_list():
    #     print(header)
    # print(json.dumps(json.loads(decompress_gzip(request.data)),indent=2))
    # json.dumps(json.loads(request.data),indent=2)

    request_headers = {}
    for key in ['Host','Accept-Encoding','Content-Length','Content-Encoding','Content-Type','User-Agent']:
        request_headers[key] = request.headers.get(key)

    insert_query = ''' INSERT INTO events(platform,type,body,headers)
              VALUES(?,?,?,?) '''
    record = ('python', 'example', request.data, json.dumps(request_headers)) # type(json.dumps(request_headers)) <type 'str'>

    try:
        with sqlite3.connect(database) as db:
            cursor = db.cursor()
            cursor.execute(insert_query, record)
            print('> sqlite3 ID', cursor.lastrowid)
            cursor.close()
    except Exception as err:
        print("LOCAL EXCEPTION SAVE", err)

    try:
        response = http.request(
            "POST", str(SENTRY), body=request.data, headers=request_headers 
        )
        return 'success'
    except Exception as err:
        print('LOCAL EXCEPTION FORWARD', err)

""" This is only for using the proxy to forward events directly to Sentry and NOT save them in your database
If you're not using this, you can ignore it
Must pass auth key in URL (not request headers) or else 403 CSRF error from Sentry
AM Transactions can't be sent to any self-hosted Sentry instance as of 10.0.0 05/30/2020 
"""
def sentryUrl(DSN):
    if ("@localhost:" in DSN):
        KEY = DSN.split('@')[0][7:]
        # assumes single-digit projectId right now
        PROJECT_ID= DSN[-1:]
        HOST = 'localhost:9000'
        return "http://%s/api/%s/store/?sentry_key=%s&sentry_version=7" % (HOST, PROJECT_ID, KEY)
    if ("ingest.sentry.io" in DSN):
        KEY = DSN.split('@')[0][8:] # 8 because of 's' in 'https'
        HOST = DSN.split('@')[1].split('/')[0]
        PROJECT_ID = DSN.split('@')[1].split('/')[1] 
        return "https://%s/api/%s/store/?sentry_key=%s&sentry_version=7" % (HOST, PROJECT_ID, KEY)
    else:
        print('\n else')
        KEY = DSN.split('@')[0][8:]
        HOST = DSN.split('@')[1].split('/')[0]
        PROJECT_ID = DSN.split('@')[1].split('/')[1] 
        return "https://%s/api/%s/store/?sentry_key=%s&sentry_version=7" % (HOST, PROJECT_ID, KEY)

def sentryUrlEnvelope(DSN):
    if ("@localhost:" in DSN):
        KEY = DSN.split('@')[0][7:]
        # assumes single-digit projectId right now
        PROJECT_ID= DSN[-1:]
        HOST = 'localhost:9000'
        return "http://%s/api/%s/envelope/?sentry_key=%s&sentry_version=7" % (HOST, PROJECT_ID, KEY)
    if ("ingest.sentry.io" in DSN):
        KEY = DSN.split('@')[0][8:] # 8 because of 's' in 'https'
        HOST = DSN.split('@')[1].split('/')[0]
        PROJECT_ID = DSN.split('@')[1].split('/')[1] 
        return "https://%s/api/%s/envelope/?sentry_key=%s&sentry_version=7" % (HOST, PROJECT_ID, KEY)
    else:
        print('\n else')
        KEY = DSN.split('@')[0][8:]
        HOST = DSN.split('@')[1].split('/')[0]
        PROJECT_ID = DSN.split('@')[1].split('/')[1] 
        return "https://%s/api/%s/envelope/?sentry_key=%s&sentry_version=7" % (HOST, PROJECT_ID, KEY)



# print(json.dumps(request_headers,indent=2))

# regular:
# print(json.dumps(json.loads(decompress_gzip(request.data)),indent=2))
# sessions:
# print(json.dumps(decompress_gzip(request.data),indent=2))
