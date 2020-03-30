import argparse
import os
import sentry_sdk
import requests
# from dotenv import load_dotenv
# load_dotenv()

# TODO attempt the workflow using platform sdk's other than python. 

# make sentry_sdk send the event to :3001 which is a Flask API and not Sentry.io
# saves in db
MODIFIED_DSN_SAVE = 'http://759bf0ad07984bb3941e677b35a13d2c@localhost:3001/2'

# saves in db and forwards the event to your Sentry instance's endpoint doesnt save but sends?
MODIFIED_DSN_SAVE_AND_FORWARD = 'http://759bf0ad07984bb3941e677b35a13d2c@localhost:3001/3'

# The proxy forwards it on through to Sentry. Doesn' save to DB
MODIFIED_DSN_FORWARD = 'http://759bf0ad07984bb3941e677b35a13d2c@localhost:3001/4'

# Unadulterated DSN, as provided by Sentry instance. As provided in the DSN Client Keys for the Organization's Project
ORIGINAL_DSN_FORWARD = 'http://759bf0ad07984bb3941e677b35a13d2c@localhost:9000/2'

def app():
    # err = raise Exception("raised exception")
    # sentry_sdk.capture_exception(err)
    raise Exception('this is the exception')

def initialize_sentry():
    params = { 'dsn': MODIFIED_DSN_SAVE }
    sentry_sdk.init(params)
    
if __name__ == '__main__':
    initialize_sentry()
    app()
