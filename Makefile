all:
	go build -o bin/main *.go && ./bin/main -n 15 err

i:
	go build -o bin/main *.go && ./bin/main -i -n 15 err
