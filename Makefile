all:
	go build -o bin/main *.go && ./bin/main -n 20 err

i:
	go build -o bin/main *.go && ./bin/main -i -n 50
