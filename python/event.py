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
DSN = "http://2ba68720d38e42079b243c9c5774e05c@o87286.ingest.sentry.io/1316515"
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
        1 / 0
    except Exception as err:
        sentry_sdk.capture_exception(err)

def app():
    # stacktrace()
    
    # Exception literals do not have stack traces
    sentry_sdk.capture_exception(Exception("Five0Ten"))

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
