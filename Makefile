all:
	go build -o bin/event-to-sentry *.go && ./bin/event-to-sentry