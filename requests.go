package main

import (
	"fmt"

	"github.com/getsentry/sentry-go"
)

var (
	counter int
)

type Requests struct {
	events []Event
	// Consider making a Constructor for these attrs, so can simplify send(). However loss of ordering of events.
	// eventsJavascript []Event
	// eventsPython []Event
}

// Doing each destination one-by-one, gives each org a rest before its API is called again, so don't insert a short Sleep Timeout yet
func (r *Requests) send() {
	for _, event := range r.events {
		// fmt.Println("\nEVENT PLATFORM", event.Platform)

		// CONSIDER check if Destinations array is empty, or do during an init somewhere
		switch event.Platform {
		case JAVASCRIPT:
			for _, dsn := range config.Destinations.Javascript {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
				counter++
			}
		case PYTHON:
			for _, dsn := range config.Destinations.Python {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
				counter++
			}
		case JAVA:
			for _, dsn := range config.Destinations.Java {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
				counter++
			}
		default:
			sentry.CaptureMessage("unsupported event platform: " + event.Platform)
			fmt.Printf("unrecognized Platform %v", event.Platform)
		}
	}
	fmt.Printf("\n> TOTAL sent: %v", counter)

	// does not Capture, not sure why
	sentry.CaptureMessage("finished sending all requests")
}
