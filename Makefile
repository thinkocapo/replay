all:
	go build -o bin/main *.go && ./bin/main

i:
	go build -o bin/main *.go && ./bin/main eventtest -i
