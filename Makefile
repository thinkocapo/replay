proxy:
	FLASK_APP=./python/proxy.py FLASK_ENV=development flask run -p 3001

event:
	go run event.go

eventpy:
	python3 python/event.py

eventtosentry:
	go run event-to-sentry.go

pythontosentry:
	python3 event-to-sentry.py

testdb:
	python3 test/db.py