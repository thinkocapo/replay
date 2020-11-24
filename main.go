package main

import (
	"flag"
	"log"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var (
	all        *bool
	ignore     *bool
	traceIds   []string
	filePrefix string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	initializeSentry()
	sentry.CaptureMessage("job started")

	// TODO - check for all needed .env vars

	ignore = flag.Bool("i", false, "ignore sending the event to Sentry.io")
	flag.Parse()

	// filePrefix of files to read
	filePrefix = os.Args[1]
}

func main() {
	demoAutomation := DemoAutomation{}

	events := demoAutomation.getEventsFromSentry()
	println("> demoAutomation events:", len(events))

	for _, event := range events {
		if event.Kind == ERROR || event.Kind == DEFAULT {
			event.Error.eventId()
			event.Error.release()
			event.Error.user()
			event.Error.timestamp()
		}
		if event.Kind == TRANSACTION {
			event.Transaction.eventId()
			event.Transaction.release()
			event.Transaction.user()
			event.Transaction.timestamps()
		}
	}

	getTraceIds(events)
	updateTraceIds(events)

	requests := Requests{events}
	requests.send()
}
