package main

import (
	"fmt"

	"github.com/getsentry/sentry-go"
)

type Requests struct {
	events []Event
	// TODO
	// eventsJavascript []Event
	// eventsPython []Event
	// ^ this would simplify the below send() function
}

func (r *Requests) send() {
	for _, event := range r.events {
		// fmt.Println("EVENT PLATFORM", event.Platform)

		switch event.Platform {
		case JAVASCRIPT:
			// CONSIDER should be a dsn, not a fullurl?
			for _, fullurl := range config.Destinations.Javascript {
				event.setDsn(fullurl)
				request := NewRequest(event)
				request.send()
			}
		case PYTHON:
			for _, fullurl := range config.Destinations.Python {
				event.setDsn(fullurl)
				request := NewRequest(event)
				request.send()
			}
		}

		// OG
		// request := NewRequest(event)
		// request.send()
	}
	fmt.Printf("> DONE sending %v events", len(r.events))

	// does not Capture, not sure why
	sentry.CaptureMessage("finished sending all requests")
}
