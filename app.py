import argparse
import os
import sentry_sdk
import requests
# from dotenv import load_dotenv
# load_dotenv()

SENTRY = 'localhost:9000'
# FLASK = 'localhost:3001'
FLASK = '0.0.0.0:3001'

# The event skips the proxy and goes directly to Sentry. DSN in its original form from Sentry
ORIGINAL_DSN = 'http://09aa0d909232457a8a6dfff118bac658@' + SENTRY + '/2'

# The proxy forwards the event on to Sentry. Doesn't save to DB
MODIFIED_DSN_FORWARD = 'http://09aa0d909232457a8a6dfff118bac658@' + FLASK + '/2'

# The proxy saves the event to database. Doesn't send to Senry.
MODIFIED_DSN_SAVE = 'http://09aa0d909232457a8a6dfff118bac658@' + FLASK + '/3'

# The proxy saves the event to database and forwards it on to Sentry
MODIFIED_DSN_SAVE_AND_FORWARD = 'http://09aa0d909232457a8a6dfff118bac658@'+ FLASK + '/4'

def app():
    sentry_sdk.capture_exception(Exception("stringobject"))
    # raise Exception('big problem')

def initialize_sentry():
    params = { 'dsn': MODIFIED_DSN_SAVE }
    sentry_sdk.init(params)
    
if __name__ == '__main__':
    initialize_sentry()
    app()
