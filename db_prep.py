from gzip import GzipFile
import json
from six import BytesIO
import sqlite3

# Functions from getsentry/sentry-python
def decompress_gzip(bytes_encoded_data):
    try:
        fp = BytesIO(bytes_encoded_data)
        try:
            print('111')
            f = GzipFile(fileobj=fp)
            print('2222')
            return f.read().decode("utf-8")
        finally:
            f.close()
    except Exception as e:
        raise e


path_to_database = r"/home/wcap/tmp/mypythonsqlite.db"
conn = sqlite3.connect(path_to_database)

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

        # READ
        cur.execute("SELECT * FROM events")
 
        rows = cur.fetchall()
        print('LENGTH', len(rows))

        for row in rows:
        # for row in cur.fetchall():
            # print('YEAAAH')
            print(row)

        # test = rows[len(rows) - 1]
        test = rows[-1]
        test = list(test)
        print("\nLast Item's ID", test[0])

        read_write_buffer = test[3]
        # print('TYPE read_write_buffer', read_write_buffer)

        # print('------------', str(read_write_buffer))
        # print('decode...', read_write_buffer.decode("utf-8"))

        json_body = decompress_gzip(str(read_write_buffer))
        dict_body = json.loads(json_body)

        print('dict_body', dict_body)

        # data = str(read_write_buffer)
        # print('TYPEOF data', type(data)) # <type 'str'> okay? is same as result as .getvalue()
        # print('TYPEOF data', type(bytearray(data))) # <type 'bytearray'>

except Exception as e:
    print(e)