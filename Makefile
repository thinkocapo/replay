proxy:
	FLASK_APP=./python/proxy.py FLASK_ENV=development flask run -p 3001
proxyhttps:
	FLASK_APP=./python/proxy.py FLASK_ENV=development flask run --cert=adhoc -p 3001

event:
	go run event.go
eventsentry:
	go run event-to-sentry.go

eventpy:
	python3 python/event.py
eventsentrypy:
	python3 event-to-sentry.py

createdb:
	python3 test/create-db.py
removedb:
	python3 test/remove-db.py
resetdb:
	removedb testdb
testdb:
	python3 test/db.py