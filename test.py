import sys

print('111111\n')

print(sys.argv)


# print('GREATER', len(args) > 1)

# args = sys.argv[1:]


_id = sys.argv[1] if len(sys.argv) > 1 else None


print('_id',_id)
if _id==None:
    print("None")
else:
    print('ID present', _id)
# print('TYPE', type(arg))
# print("arg", arg==[])