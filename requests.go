package main

import (
	"fmt"

	"github.com/getsentry/sentry-go"
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
		// fmt.Println("EVENT PLATFORM", event.Platform)

		switch event.Platform {
		case JAVASCRIPT:
			for _, dsn := range config.Destinations.Javascript {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case PYTHON:
			for _, dsn := range config.Destinations.Python {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		}
	}
	fmt.Printf("\n> DONE sending %v events", len(r.events))

	// does not Capture, not sure why
	sentry.CaptureMessage("finished sending all requests")
}
