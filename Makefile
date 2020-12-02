all:
	go build -o bin/main *.go && ./bin/main -n 5 err

i:
	go build -o bin/main *.go && ./bin/main -i -n 5 err
