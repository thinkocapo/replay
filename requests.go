package main

import (
	"fmt"

	"github.com/getsentry/sentry-go"
)

type Requests struct {
	events []Event
}

func (r *Requests) send() {
	// TODO for _, org := range orgs{ }
	for _, event := range r.events {
		request := NewRequest(event)
		// TODO for _, dsn := range destinations{ }
		request.send()
	}
	fmt.Printf("> DONE sending %v events", len(r.events))

	// does not Capture, not sure why
	sentry.CaptureMessage("finished sending all requests")
}

// TODO last
// the fact that it does each destination one-by-one, gives each a little bit of a pause, like a Sleep Timeout, so no need to code a short Sleep Timeout
// func (r *Requests) destinations() {
// 	for envvarKey, envarValue in envarPairs {
// 		switch envarKeys:
// 		case DSN_JAVASCRIPT
// 			// should be Private
// 			// for python_destinations from YAML:
// 				// destination()
// 		case DSN_PYTHON
// 	}

// }
