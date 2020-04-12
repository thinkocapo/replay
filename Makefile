all:
	flask_prep db_prep proxy

proxy:
	FLASK_APP=./flask/server-sqlite.py FLASK_ENV=development flask run -p 3001

event_to_db:
	python app.py

event_to_sentry:
# 	TODO

goreplay:
	go build middleware.go
	sudo ./gor --input-raw :9000 --middleware "./middleware" --output-http http://localhost:9000/api/2/store