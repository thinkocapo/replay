all:
	go build -o bin/main *.go && ./bin/main eventtest-s

i:
	go build -o bin/main *.go && ./bin/main eventtest-s -i
