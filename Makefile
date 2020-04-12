all:
	flask_prep db_prep proxy

flask_prep:
	cd flask && virtualenv .virtualenv && source ./flask/.virtualenv/bin/activate && pip install -r requirements.txt

db_prep:
	python sqlite-prep.py

proxy:
	FLASK_APP=./flask/server-sqlite.py FLASK_ENV=development flask run -p 3001

goreplay:
	go build middleware.go
	sudo ./gor --input-raw :9000 --middleware "./middleware" --output-http http://localhost:9000/api/2/store