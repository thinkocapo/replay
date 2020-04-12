from gzip import GzipFile
import json
from six import BytesIO
import sqlite3

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


# Full Path???
# TODO 1 try with local ./database.db
# path_to_database = r"/home/wcap/tmp/mypythonsqlite.db"

print('111111111111111')
path_to_database = r"database.db"

conn = sqlite3.connect(path_to_database)

print('2222222222222222')

sql_table_events = """ CREATE TABLE IF NOT EXISTS events (
                                        id integer PRIMARY KEY,
                                        name text,
                                        type text,
                                        data BLOB,
                                        headers BLOB
                                    ); """

try:
    with conn:
        cur = conn.cursor()

        # CREATE TABLE
        cur.execute(sql_table_events)
        print('DONE')
        
        #  TEST it worked
        # READ
        # cur.execute("SELECT * FROM events")
 
        # rows = cur.fetchall()
        # print('LENGTH', len(rows))

        # for row in rows:
        #     print(row)

        # test = rows[len(rows) - 1]
        # test = rows[-1]
        # test = list(test)
        # print('Last Item ID', test[0])

        # buffer = test[3]

        # <read-write buffer ptr 0x562a8e765e30, size 1522 at 0x562a8e765df0>
        # <type 'buffer'>
        # print('type(buffer)', type(buffer))

        # json_body = decompress_gzip(str(buffer))
        # dict_body = json.loads(json_body)

        # print('dict_body', dict_body)


except Exception as e:
    print(e)