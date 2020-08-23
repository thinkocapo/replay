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
print(database)

try:
    with open(JSON) as file:
        current_data = json.load(file)
        print(len(current_data))

    # with conn:
    #     cur = conn.cursor()
    #     rows = []
    #     _body = sys.argv[2] if len(sys.argv) > 2 else None
    #     _id = sys.argv[1] if len(sys.argv) > 1 else None
    #     if _id==None:
    #         cur.execute("SELECT * FROM events ORDER BY id;") # LIMIT 1
    #         rows = cur.fetchall()    
    #         print('TOTAL ROWS: ', len(rows))
    #     else:
    #         cur.execute("SELECT * FROM events WHERE id=?", [_id])
    #         rows = cur.fetchall()

    #     row = rows[-1]        

    #     sqlite_id = row[0]
    #     event_platform = row[1]
    #     event_type = row[2]
    #     body = row[3] # buffer
    #     headers = row[4]

    #     output = {
    #         'id': sqlite_id,
    #         'platform': event_platform,
    #         'type': event_type,
    #         'headers': json.loads(headers)
    #     }

    #     if _body == '-b':
    #         print('_body', _body)
    #         try:
    #             output['body'] = json.loads(body)
    #         except:
    #             output['body'] = body
        
    #     print(json.dumps(output, indent=2))
    
except Exception as e:
    print('EXCEPTION in test/db.py:', e)
