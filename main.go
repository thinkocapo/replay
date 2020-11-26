package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var (
	all        *bool
	ignore     *bool
	traceIds   []string
	filePrefix string
	config     Config
	n          *int
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	initializeSentry()
	sentry.CaptureMessage("job started")
	// TODO check for all other needed .env vars, besides config.yml
	// CONSIDER put all config ^ to config.yml
	ip()
	parseYaml()
	ignore = flag.Bool("i", false, "ignore sending the event to Sentry.io")
	n = flag.Int("n", 25, "default number of events to read from a source")
	flag.Parse()
	print("n is", strconv.Itoa(*n))

	// Prefix of files to read, if reading from GCS
	filePrefix = os.Args[1]
}

func main() {
	demoAutomation := DemoAutomation{}
	fmt.Println("DESTINATIONS", config.Destinations)
	events := demoAutomation.getEventsFromSentry()

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
