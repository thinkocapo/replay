proxy:
	FLASK_APP=./flask/proxy.py FLASK_ENV=development flask run -p 3001

go:
	go run event-to-sentry.go

py:
	python3 event-to-sentry.py

goreplay:
	go build middleware.go
	sudo ./gor --input-raw :9000 --middleware "./middleware" --output-http http://localhost:9000/api/2/store