proxy:
	FLASK_APP=./python/proxy.py FLASK_ENV=development flask run -p 3001
proxyhttps:
	FLASK_APP=./python/proxy.py FLASK_ENV=development flask run --cert=adhoc -p 3001

eventsentry:
	go run event-to-sentry.go
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