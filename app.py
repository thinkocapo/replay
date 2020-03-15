import argparse
from dotenv import load_dotenv
import os
import sentry_sdk
import requests
load_dotenv()
DSN = os.getenv('DSN')
DUMP_REQUEST = 'http://localhost:12345/decoderequest'

# redirects the request away from dsn and to an endpoint defined by the router in dump-request.go
def redirect(event, hint):
    # don't use b/c sends as x-www-form-urlencoded
    # r = requests.post(DUMP_REQUEST, data=event)
    
    # json
    r = requests.post(DUMP_REQUEST, json=event)
    return null

# checks cli arg for '-i'gnoring the redirect or not
def set_redirect_toggle():
    parser = argparse.ArgumentParser()
    parser.add_argument("-i", action='store_true', dest='redirect', help="ignore sending event to dsn. redirect to DumpRequest", default=False)
    args = parser.parse_args()
    params = { "dsn": DSN }
    if args.redirect == True:
        print('ignore sending request to %s. redirect request to %s' % (DSN, DUMP_REQUEST))
        params['before_send'] = redirect
    return params

def app():
    sentry_sdk.capture_exception(Exception("This is from app.py"))

if __name__ == '__main__':
    params = set_redirect_toggle()
    sentry_sdk.init(params)
    app()