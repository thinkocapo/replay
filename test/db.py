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

JSON = os.getenv('JSON')
database = JSON or os.getcwd() + "/db.json"
# print(database)

try:
    with open(JSON) as file:

        events = json.load(file)
        print(len(events))

        _body = sys.argv[2] if len(sys.argv) > 2 else None
        _id = sys.argv[1] if len(sys.argv) > 1 else None
        
        # print(events[1])       # works 
        if _id==None:
            events    
            print('TOTAL ROWS ALL: ', len(events))
        else:
            events2 = events[int(_id)]
            print('TOTAL ROWS _id: ', events2)

        # row = rows[-1]        

        # sqlite_id = row[0]
        # event_platform = row[1]
        # event_type = row[2]
        # body = row[3] # buffer
        # headers = row[4]

        # output = {
        #     'id': sqlite_id,
        #     'platform': event_platform,
        #     'type': event_type,
        #     'headers': json.loads(headers)
        # }

        # if _body == '-b':
        #     print('_body', _body)
        #     try:
        #         output['body'] = json.loads(body)
        #     except:
        #         output['body'] = body
        
        # print(json.dumps(output, indent=2))
    
except Exception as e:
    print('EXCEPTION in test/db.py:', e)
