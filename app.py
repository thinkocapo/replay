import argparse
import requests
from dotenv import load_dotenv
import os
load_dotenv()

# send to API defined by the router in dump-request.go
# do not send event to DSN Key
def redirect(event, hint):
    print('redirecting...')
    r = requests.post('http://localhost:12345/dumprequest', data=event)
    return null

import sentry_sdk
sentry_sdk.init(
    dsn=os.getenv('DSN'),
    before_send=redirect # TODO make this before_send optional (controlled by cli flag), so that running gor w/ middleware.go would pickup the request
)

def app():
    print('app me now')
    sentry_sdk.capture_exception(Exception("This is an example of an error message."))

if __name__ == '__main__':
    app()