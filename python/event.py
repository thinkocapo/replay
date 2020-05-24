import argparse
from dotenv import load_dotenv
load_dotenv()
import os
import sentry_sdk
import socket

# Unmodified DSN will send event directly to Sentry
# https://<key>@<organization>.ingest.sentry.io/<project>
# http://09aa0d909232457a8a6dfff118bac658@localhost:9000/2
DSN = os.getenv('DSN')

# send event to Sentry or the Flask proxy which interfaces with Sqlite
KEY = DSN.split('@')[0]
PROXY = 'localhost:3001'
# SENTRY = 'localhost:9000'

# proxy forwards the event on to Sentry. Doesn't save to DB
MODIFIED_DSN_FORWARD = KEY + '@' + PROXY + '/2'

# proxy saves the event to database. Doesn't send to Senry.
MODIFIED_DSN_SAVE = KEY + '@' + PROXY + '/3'

# proxy saves the event to database and forwards it on to Sentry
MODIFIED_DSN_SAVE_AND_FORWARD = KEY + '@'+ PROXY + '/4'

# def dsn(string):
#     return KEY + '@'+ PROXY + string
# MODIFIED_DSN_FORWARD = dsn('/2')
# MODIFIED_DSN_SAVE = dsn('/3')
# MODIFIED_DSN_SAVE_AND_FORWARD = dsn('/4')

def stacktrace():
    try:
        1 / 0
    except Exception as err:
        # sentry_sdk.capture_exception(err)
        sentry_sdk.capture_exception(Exception("things1022"))

def app():
    stacktrace()
    # sentry_sdk.capture_exception(Exception("This won't have a stack trace"))

def dsn_and_proxy_check():
    if DSN=='':
        print('> no DSN')
        exit()
    try:
        HOST,PORT = PROXY.split(':')
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.connect((HOST, int(PORT)))
    except Exception as exception:
        print("> proxy not running")
    finally:
        s.close()
    
def initialize_sentry():
    params = { 'dsn': MODIFIED_DSN_SAVE }
    sentry_sdk.init(params)
    
if __name__ == '__main__':
    dsn_and_proxy_check()
    initialize_sentry()
    app()
