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

# if path is outside of directory, must use absolute path like /home/user/database.db
# path_to_database = r"sqlite.db"
path_to_database = r"am-transactions-sqlite.db"

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
        cur.execute("SELECT * FROM events")
        rows = cur.fetchall()
        print('TOTAL ROWS: ', len(rows))
    
except Exception as e:
    print(e)