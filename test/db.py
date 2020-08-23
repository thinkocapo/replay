from dotenv import load_dotenv
from gzip import GzipFile
import json
from six import BytesIO
import os
import sys
load_dotenv()

"""
This tests how many records are in your json file
`python3 test/db.py 5' gets the 5th item
`python3 test/dby.py 5 -b' gets the 5th item and prints its body
Otherwise the total row count and most recently saved item is printed
"""

# Functions from getsentry/sentry-python
def decompress_gzip(bytes_encoded_data):
    try:
        fp = BytesIO(bytes_encoded_data)
        try:
            f = GzipFile(fileobj=fp)
            return f.read().decode("utf-8")
        finally:
            f.close()
    except Exception as e:
        raise e

database = os.getenv('JSON') or os.getcwd() + "/db.json"

try:
    with open(database) as file:

        events = json.load(file)
        event = {}

        _body = sys.argv[2] if len(sys.argv) > 2 else None
        _id = sys.argv[1] if len(sys.argv) > 1 else None
        
        if _id==None:
            events
            event = events[0]    
            print('len(events)', len(events))
        else:
            event = events[int(_id)]
            print('selecting 1 event...')

        sqlite_id = _id
        event_platform = event['platform']
        event_type = event['kind']
        body = event['body']
        headers = event['headers']

        output = {
            'id': _id,
            'platform': event['platform'],
            'type': event['kind'],
            'headers': event['headers']
            # 'headers': json.loads(headers)
        }

        if _body == '-b':
            print('_body', _body)
            try:
                output['body'] = json.loads(body)
            except:
                output['body'] = body
        
        print(json.dumps(output, indent=2))
    
except Exception as e:
    print('EXCEPTION in test/db.py:', e)
