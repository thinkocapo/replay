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

        print('type is', test[3].read().hex())



except Error as e:
    print(e)