package main

import (

	// "encoding/json"
	"flag"

	// "io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	// "google.golang.org/api/iterator"
)

var httpClient = &http.Client{} // TODO

var (
	all    *bool
	id     *string
	ignore *bool
	db     *string // or similar could be for dynamically handling the 'prefix' for Bucket
	js     *string
	py     *string
	// SENTRY_URL  string
	projectDSNs map[string]*DSN
	traceIds    []string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	all = flag.Bool("all", false, "send all events. default is send latest event")
	// id = flag.String("id", "", "id of event in sqlite database") // 08/27 non-functional today
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

const ERROR = "error"
const TRANSACTION = "transaction"

func main() {

	demoAutomation := DemoAutomation{}

	events := demoAutomation.getEvents()

	for _, event := range events {
		if event.Kind == ERROR { // TODO ERROR const
			event.Error.eventId()
			event.Error.release()
			event.Error.user()
			event.Error.timestamp()
		}
		if event.Kind == "transaction" { // TODO TRANSACTION const
			event.Transaction.eventId()
			event.Transaction.release()
			event.Transaction.user()
			event.Transaction.timestamps()
			// TODO .measurements()
		}
	}

	getTraceIds(events)
	updateTraceIds(events)

	requests := Requests{events}
	requests.send()

	return
}

// demoAutomation.init()

// TODO double check it's object was updated reference `fmt.Println("\n> timestamp AFTER", event.Error.Timestamp)`
