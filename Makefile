proxy:
	FLASK_APP=./python/proxy.py FLASK_ENV=development flask run -p 3001

proxyhttps:
	FLASK_APP=./python/proxy.py FLASK_ENV=development flask run --cert=adhoc -p 3001

event:
	go run event.go

eventpy:
	python3 python/event.py

eventsentry:
	go run event-to-sentry.go

eventsentrypy:
	python3 event-to-sentry.py

resetdb:
	rm tracing-example.db && touch tracing-example.db && python3 test/create-db.py
testdb:
	python3 test/db.py