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
path_to_database = r"sqlite.db"
conn = sqlite3.connect(path_to_database)

try:
    with conn:
        cur = conn.cursor()
        cur.execute("SELECT * FROM events")

        rows = cur.fetchall()
        print('TOTAL ROWS: ', len(rows))
 
        row = rows[-1] # -1 is python event right now
        # row = list(row)
        print('ID OF LATEST ROW', row[0]) # is latest??

        # <read-write buffer ptr 0x562a8e765e30, size 1522 at 0x562a8e765df0>
        # <type 'buffer'>
        buffer = row[3]

        headers = row[4]
        print('\nHEADERS\n', headers)



    
        # TODO why did this stop working "EXCEPTION a bytes-like object is required, not 'str'"
        print('type(buffer)', type(buffer))

        # TODO - this errors on javascript, which might not be gzip'd
        
        # python, works
        # json_body = decompress_gzip(buffer)
        # dict_body = json.loads(json_body)

        # javascript, works
        dict_body = json.loads(buffer)

        print('dict_body', dict_body['event_id'])
        # print('dict_body', dict_body['timestamp']) # fails in javascript

        # print(dict_body)
        for key in dict_body:
            print(key, type(dict_body[key]))
except Exception as e:
    print('EXCEPTION', e)