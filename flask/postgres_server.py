# 04/08/2020 from master

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

    with db.connect() as conn:
        conn.execute(insert_query, record)
        conn.close()
    
    print("created event in postgres")
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

    # the try/except here has not been tested yet
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

    if pk==0:
        query = "SELECT * FROM events ORDER BY pk DESC LIMIT 1;"
    else:
        query = "SELECT * FROM events WHERE pk={};".format(pk)
    
    with db.connect() as conn:
        rows = conn.execute(query).fetchall()
        conn.close()
        # <class 'sqlalchemy.engine.result.RowProxy'
        row_proxy = rows[0]
 
    # row_proxy.data is <class bytes> so row_proxy.data is b'\x1f\x8b\
    json_body = decompress_gzip(row_proxy.data)

    # update event_id/timestamp so Sentry will accept the event again
    dict_body = json.loads(json_body)
    dict_body['event_id'] = uuid.uuid4().hex
    dict_body['timestamp'] = datetime.datetime.utcnow().isoformat() + 'Z'

    bytes_io_body = compress_gzip(dict_body)
    
    try:
        # bytes_io_body.getvalue() is for reading the bytes
        response = http.request(
            "POST", str(SENTRY_API_STORE_ONPREMISE), body=bytes_io_body.getvalue(), headers=row_proxy.headers 
        )
    except Exception as err:
        print('LOCAL EXCEPTION', err)

    return 'loaded and forwarded to Sentry'