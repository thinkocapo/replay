# Deprecated
# initialize_sentry...
# parser = argparse.ArgumentParser()
# parser.add_argument("-r", action='store_true', dest='redirect', help="ignore sending event to dsn. redirect to a homemade API", default=False)
# args = parser.parse_args()
# if args.redirect == True:
#     params['before_send'] = before_send_redirect
# ...
# def before_send_redirect(event, hint):
#     try:
#         r = requests.post(DUMP_REQUEST, json=event)
#         return event
#     except Exception as err:
#         print(err)
#         return 'failed'
#     return null

##########################################################################################

# data = decompress_gzip(request.data)
        # print('type(data)', type(data)) # <class 'str'>...
        # print('data', data) # {"exception": {"values": [{"stacktrace": {"...
        # body = io.BytesIO()
        # with gzip.GzipFile(fileobj=body, mode="w") as f:
        #     f.write(json.dumps(data, allow_nan=False).encode("utf-8"))

        # NOTE body=body.getvalue() errors in the onprem Sentry as "b'{"error":"Bad data decoding request (TypeError, Incorrect padding)"}'"

# Got error w/ 403-csrf.html until I put X-Sentry-Auth in URL rather than headers, which then gave error on the onprem Internal project
# But (my interpretation of) getsentry/sentry-python shows it being set in the request's headers, not URL.
    # 'X-Sentry-Auth': headers.get('X-Sentry-Auth'),
    # 'X-Sentry-Auth': 'Sentry sentry_key=759bf0ad07984bb3941e677b35a13d2c, sentry_version=7, sentry_client=sentry.python/0.14.2',
    
# def decompress_gzip(encoded_data):
#     try:
#         fp = BytesIO(encoded_data)
#         try:
#             f = GzipFile(fileobj=fp)
#             return f.read().decode("utf-8")
#         finally:
#             f.close()
#     except Exception as e:
#         raise e

# def get_connection():
#     with sentry_sdk.start_span(op="psycopg2.connect"):
#         connection = psycopg2.connect(
#             host=HOST,
#             database=DATABASE,
#             user=USERNAME,
#             password=PASSWORD)
#     return connection
###########################################################################################


# connection = get_connection()
# cursor = connection.cursor(cursor_factory = psycopg2.extras.DictCursor)
# try:
#     cursor.execute(insert_query, (name, tool_type, randomString(10), image, random.randint(10,50)))
#     connection.commit()
# except:
#     raise "Row insert failed\n"
#     return 'fail'
# cursor.close()
# connection.close()

# rows = []
# for row in results:
#     rows.append(dict(row))
# return json.dumps(rows)