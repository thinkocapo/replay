package main

import (
	"flag"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var (
	sentry   interface{} 
	sentry "github.com/getsentry/sentry-go"
	sentry   sentry-go
	all      *bool
	ignore   *bool
	DSNs     map[string]*DSN
	traceIds []string
	xx       interface{}
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	ignore = flag.Bool("i", false, "ignore sending the event to Sentry.io")
	flag.Parse()
}

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://d732e60c53c842409b5900b58d005f25@o87286.ingest.sentry.io/1507371", // os.Getenv("SENTRY"),
		// Either set environment and release here or set the SENTRY_ENVIRONMENT
		// and SENTRY_RELEASE environment variables.
		// Environment: "",
		// Release:     "",
		// Enable printing of SDK debug messages.
		// Useful when getting started or trying to figure something out.

		// TEST
		Debug: true,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	// Set the timeout to the maximum duration the program can afford to wait.
	defer sentry.Flush(2 * time.Second)
	sentry.CaptureMessage("It works!")

	demoAutomation := DemoAutomation{}

	events := demoAutomation.getEvents()

	for _, event := range events {
		if event.Kind == ERROR {
			event.Error.eventId()
			event.Error.release()
			event.Error.user()
			event.Error.timestamp()
		}
		if event.Kind == "transaction" {
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
