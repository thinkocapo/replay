from dotenv import load_dotenv
from gzip import GzipFile
import json
from six import BytesIO
import os
import sqlite3
import sys
load_dotenv()

"""
This tests how many records are in your Sqlite database
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

# if path is outside of directory, must use absolute path like /home/user/database.db
# path_to_database = r"sqlite.db"
# path_to_database = r"am-transactions-sqlite.db"
# path_to_database = r"tracing-example.db"
SQLITE = os.getenv('SQLITE')
database = SQLITE or os.getcwd() + "/sqlite.db"
print(database)

conn = sqlite3.connect(database)

try:
    with conn:
        cur = conn.cursor()
        rows = []

        _body = sys.argv[2] if len(sys.argv) > 2 else None
        _id = sys.argv[1] if len(sys.argv) > 1 else None
        if _id==None:
            cur.execute("SELECT * FROM events ORDER BY id;") # LIMIT 1
            rows = cur.fetchall()    
            print('TOTAL ROWS: ', len(rows))
        else:
            cur.execute("SELECT * FROM events WHERE id=?", [_id])
            rows = cur.fetchall()

        row = rows[-1]        
        
        # <read-write buffer ptr 0x562a8e765e30, size 1522 at 0x562a8e765df0>
        # <type 'buffer'>

        sqlite_id = row[0]
        event_name = row[1]
        event_type = row[2]
        body = row[3] # buffer
        headers = row[4]

        # TODO add flag for 'include body' in query
        output = {
            'id': sqlite_id,
            'platform': event_name,
            'type': event_type,
            'headers': json.loads(headers)
        }
        if _body == '-b':
            print('_body', _body)
            output['body'] = json.loads(body)
        print(json.dumps(output, indent=2))
    
except Exception as e:
    print('EXCEPTION test/db.py:', e)

# for key in dict_body:
#     print(key, type(dict_body[key]))

# in previous versions, saved bytes vs. gzipped bytes. Today, this should always be same
# print('type(body)', type(body))

# old for python
# json_body = decompress_gzip(body)
# dict_body = json.loads(json_body)

# javascript, works
# dict_body = json.loads(body)
# print('dict_body', dict_body['event_id'])
# print('dict_body', dict_body['timestamp']) # different timestamp formats depending js vs python
