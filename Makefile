all:
	db_prep flask_prep proxy

db_prep:
	python flask/sqlite-prep.py

flask_prep:
	cd flask && pip install -r requirements.txt

proxy:
	FLASK_APP=./flask/server-sqlite.py FLASK_ENV=development flask run -p 3001

goreplay:
	go build middleware.go
	sudo ./gor --input-raw :9000 --middleware "./middleware" --output-http http://localhost:9000/api/2/store