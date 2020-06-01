from gzip import GzipFile
import json
from six import BytesIO
import sqlite3

"""This tests how many records are in your Sqlite database"""

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
path_to_database = r"am-transactions-sqlite.db"

conn = sqlite3.connect(path_to_database)

try:
    with conn:
        cur = conn.cursor()
        cur.execute("SELECT * FROM events")

        rows = cur.fetchall()
        print('TOTAL ROWS: ', len(rows))
        # print('Most recent sqlite id:', rows[len(rows)-1][0]) # is latest??
 
        # TODO - iterate through all rows and print
        row = rows[-1]
        # row = list(row)

        # <read-write buffer ptr 0x562a8e765e30, size 1522 at 0x562a8e765df0>
        # <type 'buffer'>

        sqlite_id = row[0]
        event_name = row[1]
        event_type = row[2]
        buffer = row[3]
        headers = row[4]

        output = {
            'id': sqlite_id,
            'platform': event_name,
            'type': event_type
            # 'headers': headers
        }
        print(json.dumps(output, indent=2))
    
        # logs different depending on if you saved bytes or gzipped bytes
        print('type(buffer)', type(buffer))

        # old for python
        # json_body = decompress_gzip(buffer)
        # dict_body = json.loads(json_body)

        # javascript, works
        # dict_body = json.loads(buffer)
        # print('dict_body', dict_body['event_id'])
        # print('dict_body', dict_body['timestamp']) # different timestamp formats depending js vs python

        # for key in dict_body:
        #     print(key, type(dict_body[key]))
except Exception as e:
    print('EXCEPTION test/db.py:', e)