package main

import (
	"flag"
	"github.com/getsentry/sentry-go"
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
	platforms  []string
)

const JAVASCRIPT = "javascript"
const PYTHON = "python"
const JAVA = "java"
const RUBY = "ruby"
const GO = "go"
const PHP = "php"
const NODE = "node"
const DART = "dart"
const CSHARP = "csharp"
const ELIXIR = "elixir"
const PERL = "perl"
const RUST = "rust"
const COCOA = "cocoa"
const ANDROID = "android"
const FLUTTER = "flutter"
const CORDOVA = "cordova"

func init() {
	parseYamlConfig()

	initializeSentry()
	sentry.CaptureMessage("job started")

	ignore = flag.Bool("i", false, "ignore sending the event to Sentry.io")
	n = flag.Int("n", 25, "default number of events to read from a source")

	defaultPrefix := "error"
	filePrefix = flag.String("prefix", defaultPrefix, "file prefix")
	flag.Parse()
	platforms = []string{JAVASCRIPT, PYTHON, JAVA, RUBY, GO, NODE, PHP, CSHARP, DART, ELIXIR, PERL, RUST, COCOA, ANDROID, FLUTTER, CORDOVA}
}

func main() {
	demoAutomation := DemoAutomation{}

	events := demoAutomation.getEventsFromGCS()

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
