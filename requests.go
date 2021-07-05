package main

import (
	"fmt"

	"github.com/getsentry/sentry-go"
)

type Requests struct {
	events []Event
}

// TODO in each case met, check if len(config.Destinations.Platform) != 0
// Looping through each destination (DSN) sequentially gives the DSN/org's API a brief rest. the sleep timesouts are in request.go
func (r *Requests) send() {
	var found bool
	for _, event := range r.events {
		for _, platform := range platforms {
			found = false
			if platform == event.Platform {
				found = true
				for _, dsn := range config.Destinations[platform] {
					event.setDsn(dsn)
					request := NewRequest(event)
					request.send()
				}
				break
			}
		}
		if found == false {
			sentry.CaptureMessage("unsupported event platform: " + event.Platform)
			fmt.Printf("\nunrecognized Platform %v\n", event.Platform)
		}
	}
	fmt.Printf("\n> TOTAL sent: %v", counter)

	// Does not Capture, not sure why
	sentry.CaptureMessage("finished sending all requests")
}
