from dotenv import load_dotenv
from gzip import GzipFile
import json
from six import BytesIO
import os
import sqlite3
load_dotenv()

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

SQLITE = os.getenv('SQLITE')
database = SQLITE or os.getcwd() + "/sqlite.db"

print("> database", database)

conn = sqlite3.connect(database)

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