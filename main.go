package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var (
	all         *bool
	id          *string
	ignore      *bool
	eventIds    *string // TODO from a config file, for which eventIds JSON URL's to call
	js          *string
	py          *string
	projectDSNs map[string]*DSN
	traceIds    []string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	all = flag.Bool("all", false, "send all events")
	ignore = flag.Bool("i", false, "ignore sending the event to Sentry.io")
	js = flag.String("js", "", "javascript DSN")
	py = flag.String("py", "", "python DSN")
	flag.Parse()

	projectDSNs = make(map[string]*DSN)
	projectDSNs["javascript"] = parseDSN(os.Getenv("DSN_JAVASCRIPT_SAAS"))
	if *js != "" {
		projectDSNs["javascript"] = parseDSN(*js)
	}
	projectDSNs["python"] = parseDSN(os.Getenv("DSN_PYTHON_SAAS"))
	if *py != "" {
		projectDSNs["python"] = parseDSN(*py)
	}
}

func main() {

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

	return
}
