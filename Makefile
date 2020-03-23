all:
	FLASK_APP=./flask/server.py flask run -p 3001

gore:
	go build middleware.go
	sudo ./gor --input-raw :9000 --middleware "./middleware" --output-http http://localhost:9000/api/2/store