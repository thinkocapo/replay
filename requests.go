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
		// switch event.Platform
		// case JAVASCRIPT
		// for _, dsn := range destinations.JAVASCRIPT
		// event.set(dsn)
		// request.send()
		// case PYTHON
		// for _, dsn := range destinations.PYTHON
		// event.set(dsn)
		// request.send()
		request := NewRequest(event)
		request.send()
	}
	fmt.Printf("> DONE sending %v events", len(r.events))

	// does not Capture, not sure why
	sentry.CaptureMessage("finished sending all requests")
}
