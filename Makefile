all:
	go build -o bin/main *.go && ./bin/main -n 50

i:
	go build -o bin/main *.go && ./bin/main -i -n 20
