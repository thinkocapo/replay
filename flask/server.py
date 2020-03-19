import os
from flask import Flask, request, json, abort
from flask_cors import CORS
import requests
import sentry_sdk
from sentry_sdk.integrations.flask import FlaskIntegration

DSN = "https://18562a9e8e3943088b1ca3cedf222e21@sentry.io/1435220"
DSN_HTTP = "https://18562a9e8e3943088b1ca3cedf222e21@sentry.io/api/1435220/store/"
 
sentry_sdk.init(
    dsn=DSN,
    integrations=[FlaskIntegration()],
    release=os.environ.get("VERSION")
)

app = Flask(__name__)
CORS(app)

@app.route('/api/42/store/', methods=['POST'])
def event():
    event = json.loads(request.data)
    try:
        for key in event:
            print("KEY", key)
        print("to: ", DSN_HTTP)
        r = requests.post(DSN_HTTP, json=event)

        # 1
        # urllib3 - see if i can construct the right headers

        # 2 
        # convert to bytes and save in db
        # 3
        # load from db, and pass to sentry_exception in step#3

        # 3
        # only wants an Exception passed to it, not an event.
        # could set the server_name, extra, tags, breadcrumbs and things manually on it, event_id, timestamp?
        # would it miss the original request headers thoughhh?
        # sentry_sdk.capture_exception(event) # NO, wants event object ONLY and not headers+body
    except Exception as err:
        print(err)
        return 'failed'
    return 'Success - handled'

@app.route('/handled', methods=['GET'])
def handled_exception():
    # event = json.loads(request.data)
    try:
        # print(obj)
        print("HELLO")
    except Exception as err:
        return 'failed'

    return 'Success - handled'

@app.route('/unhandled', methods=['POST'])
def unhandled_exception():
    obj = {}
    obj['keyDoesntExist']

# @app.before_request
# def sentry_event_context():

#     if (request.data):
#         order = json.loads(request.data)
#         with sentry_sdk.configure_scope() as scope:
#                 scope.user = { "email" : order["email"] }
        
#     transactionId = request.headers.get('X-Transaction-ID')
#     sessionId = request.headers.get('X-Session-ID')
#     global Inventory

#     with sentry_sdk.configure_scope() as scope:
#         scope.set_tag("transaction_id", transactionId)
#         scope.set_tag("session-id", sessionId)
#         scope.set_extra("inventory", Inventory)

# @app.route('/checkout', methods=['POST'])
# def checkout():

#     order = json.loads(request.data)
#     print "Processing order for: " + order["email"]
#     cart = order["cart"]
    
#     process_order(cart)

#     return 'Success'
