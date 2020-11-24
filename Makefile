all:
	go build -o bin/main *.go && ./bin/main eventtest

i:
	go build -o bin/main *.go && ./bin/main -i eventtest
