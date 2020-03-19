import argparse
# from dotenv import load_dotenv
import os
import sentry_sdk
import requests
# load_dotenv()

# sentry_sdk will not reject this, and will send it to the Flask API of localhost:3001/42
MODIFIED_DSN = 'http://2ba68720d38e42079b243c9c5774e05c@localhost:3001/42'

# set the (optional) redirect
DUMP_REQUEST = os.getenv('DUMP_REQUEST')
FLASK_API = 'http://localhost:3001'
DSN_REDIRECT = FLASK_API

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
        # optional
        r = requests.post(DSN_REDIRECT, json=event) # DSN_REDIRECT is going to be the FLASK app or GO app (DUMP_REQUEST)
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
    params['dsn'] = DSN
    print(params)
    sentry_sdk.init(params)
    app()