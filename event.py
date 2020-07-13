import argparse
from dotenv import load_dotenv
load_dotenv()
import os
import sentry_sdk
import socket

""" 
DSN Declarations:
https://<key>@<organization>.ingest.sentry.io/<project>
https://<key>@<organization>sentry.io/<project>
https://<key>@localhost:9000/<project>

An original (unmodified) DSN will send event directly to Sentry
SDK requires that DSN ends in a number. Zero's didn't work.
# SENTRY_SELF_HOSTED = 'localhost:9000'

Could do 'python event.py -s for "save" or -f for "forward" or -sf for "save_and_forward"
"""

# send event to Sentry or the Flask proxy which interfaces with Sqlite
# DSN = os.getenv('DSN_PYTHONTEST')
DSN = "https://b9cd20b63679421e8edfea05ab1c0a06@o87286.ingest.sentry.io/5331257"
KEY = DSN.split('@')[0]
PROXY = 'localhost:3001'


# proxy forwards the event on to Sentry. Doesn't save to DB
MODIFIED_DSN_FORWARD = KEY + '@' + PROXY + '/2'

# proxy saves the event to database. Doesn't send to Senry.
MODIFIED_DSN_SAVE = KEY + '@' + PROXY + '/3'

# proxy saves the event to database and forwards it on to Sentry
MODIFIED_DSN_SAVE_AND_FORWARD = KEY + '@'+ PROXY + '/4'

def stacktrace():
    try:
        o = {}
        o['nowheretoebfound']
    except Exception as err:
        print('excepptioning...')
        sentry_sdk.capture_exception(err)

def app():
    stacktrace()
    
    # Exception literals do not have stack traces
    # sentry_sdk.capture_exception(Exception("Five0Ten"))

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
    params = { 'dsn': DSN, 'release': 'WILL.0.6' }
    sentry_sdk.init(params)
    
if __name__ == '__main__':
    # dsn_and_proxy_check()
    initialize_sentry()
    app()
