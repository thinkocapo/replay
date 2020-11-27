all:
	go build -o bin/main *.go && ./bin/main -n 20

i:
	go build -o bin/main *.go && ./bin/main -i -n 20
