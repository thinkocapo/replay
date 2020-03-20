import argparse
# from dotenv import load_dotenv
import os
import sentry_sdk
import requests
# load_dotenv()

# trick the sentry_sdk into sending the event to the Flask API on localhost:3001/42
MODIFIED_DSN = 'http://18562a9e8e3943088b1ca3cedf222e21@localhost:3001/42'

# this redirect is optional
DUMP_REQUEST = os.getenv('DUMP_REQUEST')


def app():
    with sentry_sdk.configure_scope() as scope:
        scope.set_tag("customer", "special")
    sentry_sdk.capture_exception(Exception("This is from app.py."))

def initialize_sentry():
    params = set_redirect_toggle()
    params['dsn'] = MODIFIED_DSN
    print(params)
    sentry_sdk.init(params)

# checks cli arg for '-r' for redirecting the event
def set_redirect_toggle():
    parser = argparse.ArgumentParser()
    parser.add_argument("-r", action='store_true', dest='redirect', help="ignore sending event to dsn. redirect to DumpRequest", default=False)
    args = parser.parse_args()
    params = {}
    if args.redirect == True:
        params['before_send'] = before_send_redirect
    return params

# redirects to dump-request.go or to whatever VARIABLE you define
def before_send_redirect(event, hint):
    '''returning the event is optional'''
    try:
        # optional
        r = requests.post(DUMP_REQUEST, json=event)
        return event
    except Exception as err:
        print(err)
        return 'failed'
    return null
    
if __name__ == '__main__':
    initialize_sentry()
    app()