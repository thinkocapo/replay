import argparse
import os
import sentry_sdk
import requests
# from dotenv import load_dotenv
# load_dotenv()

# make sentry_sdk send the event to :3001 which is a Flask API and not Sentry.io
MODIFIED_DSN = 'http://759bf0ad07984bb3941e677b35a13d2c@localhost:3001/2'

def app():
    sentry_sdk.capture_exception(Exception("middleman_15"))

def initialize_sentry():
    params = { 'dsn': MODIFIED_DSN }
    sentry_sdk.init(params)
    
if __name__ == '__main__':
    initialize_sentry()
    app()


# initialize_sentry...
# parser = argparse.ArgumentParser()
# parser.add_argument("-r", action='store_true', dest='redirect', help="ignore sending event to dsn. redirect to a homemade API", default=False)
# args = parser.parse_args()
# if args.redirect == True:
#     params['before_send'] = before_send_redirect

# redirects to what's defined in DUMP_REQUEST...
# def before_send_redirect(event, hint):
#     try:
#         r = requests.post(DUMP_REQUEST, json=event)
#         return event
#     except Exception as err:
#         print(err)
#         return 'failed'
#     return null