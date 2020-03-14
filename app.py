import argparse
from dotenv import load_dotenv
import os
load_dotenv()

import sentry_sdk
# sentry_sdk.init(dsn=os.getenv('DSN'))
# sentry_sdk.init(dsn='localhost:8000') #fails
# sentry_sdk.init(dsn='http://foo:bar@127.0.0.1:12345/1') # should have worked?

sentry_sdk.init("http://759bf0ad07984bb3941e677b35a13d2c@localhost:9000/2")


def app():
    print('app me now')
    sentry_sdk.capture_exception(Exception("This is an example of an error message."))

if __name__ == '__main__':
    app()