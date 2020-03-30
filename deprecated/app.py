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