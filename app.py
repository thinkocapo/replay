import argparse
import os
import sentry_sdk

from dotenv import load_dotenv
load_dotenv()

# DSN is like "https://<key>@<organization>.ingest.sentry.io/<project>"
DSN = os.getenv('DSN')

# send event to Sentry or the Flask proxy which interfaces with Sqlite
KEY = DSN.split('@')[0]
SENTRY = 'localhost:9000'
PROXY = 'localhost:3001'

# event skips the proxy and goes directly to Sentry. DSN in its original form from Sentry
ORIGINAL_DSN = KEY + '@' + SENTRY + '/2'

# proxy forwards the event on to Sentry. Doesn't save to DB
MODIFIED_DSN_FORWARD = KEY + '@' + PROXY + '/2'

# proxy saves the event to database. Doesn't send to Senry.
MODIFIED_DSN_SAVE = KEY + '@' + PROXY + '/3'

# proxy saves the event to database and forwards it on to Sentry
MODIFIED_DSN_SAVE_AND_FORWARD = KEY + '@'+ PROXY + '/4'

def app():
    sentry_sdk.capture_exception(Exception("april12226..."))

def initialize_sentry():
    params = { 'dsn': MODIFIED_DSN_SAVE }
    sentry_sdk.init(params)
    
if __name__ == '__main__':
    initialize_sentry()
    app()
