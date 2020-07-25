import gzip
from gzip import GzipFile
import io
import json
from six import BytesIO

""" Functions from getsentry/sentry-python"""
# returns json
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

def compress_gzip(dict_body):
    try:
        body = io.BytesIO()
        with gzip.GzipFile(fileobj=body, mode="w") as f:
            f.write(json.dumps(dict_body, allow_nan=False).encode("utf-8"))
    except Exception as e:
        raise e
    return body

# Using the 'transaction' property because event.type is not set uniformly across js/python and both errors+transactions
# TODO check if it's gzipped, so then could remove 'platform' parameter
def get_event_type(bytes_data, platform):
    body_dict = ''
    if platform == 'python':
        body_dict = json.loads(decompress_gzip(bytes_data))
    if platform == 'javascript':
        body_dict = json.loads(bytes_data)
    if platform == 'android':
        try:
            body_dict = json.loads(decompress_gzip(bytes_data))
        except:
            print('it is a session', decompress_gzip(bytes_data))
            result = 'session'
            return result
    
    result = ''
    if 'exception' in body_dict:
        result = 'error'
    else:
        result = 'transaction'

    return result
    # if "type" in body_dict:
    #     print("> type ", body_dict['type'])
    # if "transaction" in body_dict:
    #     print("> transaction ", body_dict['transaction'])
    # print(json.dumps(body_dict, indent=2))

    # if 'type' in body_dict:
    #     result = 'error'
    # if 'transaction' not in body_dict:
    #     result = 'error'