all:
	go build -o bin/main *.go && ./bin/main -n 5

i:
	go build -o bin/main *.go && ./bin/main -i -n 2
