package main

import "github.com/getsentry/sentry-go"

type Requests struct {
	events []Event
}

func (r *Requests) send() {
	for _, event := range r.events {
		request := NewRequest(event)
		request.send()
	}
	// does not Capture, not sure why
	sentry.CaptureMessage("finished sending all requests")
}
