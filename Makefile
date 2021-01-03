all:
	go build -o bin/main *.go && ./bin/main -n 100 -prefix err

i:
	go build -o bin/main *.go && ./bin/main -i -n 100 -prefix err
