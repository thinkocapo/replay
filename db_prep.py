import sqlite3

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
    c = conn.cursor()
    c.execute(sql_table_events)
except Error as e:
    print(e)