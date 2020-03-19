import argparse
from dotenv import load_dotenv
import os
import sentry_sdk
import requests
load_dotenv()

# senter_sdk initialization
DSN = os.getenv('DSN')

# sentry_sdk for redirects
DSN_FLASK = os.getenv('DSN_FLASK')
DUMP_REQUEST = os.getenv('DUMP_REQUEST')

# set the redirect
DSN_REDIRECT = DSN_FLASK

# checks cli arg for '-r' for redirecting the event
def set_redirect_toggle():
    parser = argparse.ArgumentParser()
    parser.add_argument("-r", action='store_true', dest='redirect', help="ignore sending event to dsn. redirect to DumpRequest", default=False)
    args = parser.parse_args()
    params = {}
    if args.redirect == True:
        print('from: %s\n  to: %s' % (DSN, DSN_REDIRECT))
        params['before_send'] = before_send_redirect
    return params

# redirects the request away from dsn and to an endpoint defined by the router in dump-request.go
# or
# redirects the request to a homemade api and still sends to DSN
def before_send_redirect(event, hint):
    '''returning the event is optional'''
    try:
        #r = requests.post(DSN_REDIRECT, json=event) # DSN_REDIRECT is going to be the FLASK app or GO app (DUMP_REQUEST)
        # TODO save event as a []bytes, load from DB
        # TODO in sentry-python sdk, find what function is called as before_send returns
        return event
    except Exception as err:
        print(err)
        return 'failed'
    return null

def app():
    with sentry_sdk.configure_scope() as scope:
        scope.set_tag("customer", "SPECIAL")
    sentry_sdk.capture_exception(Exception("This is from app.py"))

if __name__ == '__main__':
    params = set_redirect_toggle()
    params['dsn'] = DSN_FLASK # or DSN_REDIRECT?
    print(params)
    sentry_sdk.init(params)
    app()