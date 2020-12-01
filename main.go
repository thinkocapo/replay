package main

import (
	"flag"
	"log"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var (
	all        *bool
	ignore     *bool
	traceIds   []string
	filePrefix *string
	config     Config
	n          *int
	counter    int
)

// v1.0.2
func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	initializeSentry()
	sentry.CaptureMessage("job started")

	parseEnv()
	parseYaml()

	ignore = flag.Bool("i", false, "ignore sending the event to Sentry.io")
	n = flag.Int("n", 25, "default number of events to read from a source")
	filePrefix = flag.String("prefix", "err", "file prefix")
	flag.Parse()
}

func main() {
	demoAutomation := DemoAutomation{}

	// TODO append these 2 together, or combine in a single .getEvents() method
	// events := demoAutomation.getEventsFromSentry()
	// events := demoAutomation.getEventsFromGCS(*filePrefix)
	events := demoAutomation.getEvents()

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
