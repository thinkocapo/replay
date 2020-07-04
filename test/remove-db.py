
from dotenv import load_dotenv
import os
load_dotenv()

database = os.getenv('SQLITE')

decision = input("> Remove {} are you sure? (y/n): ".format(database))

if decision == "y":
    os.remove(database)
    print('database cleared')
else:
    print('nothing happened')