package main

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"math/rand"
	"time"
)

type Requests struct {
	events []Event
}

// TODO in each case met, check if len(config.Destinations.Platform) != 0
// Doing each destination one-by-one, gives each org a rest before its API is called again, so don't insert a short Sleep Timeout yet
func (r *Requests) send() {
	var found bool
	for _, event := range r.events {
		for _, platform := range platforms {
			found = false
			if platform == event.Platform {
				found = true
				for _, dsn := range config.Destinations[platform] {

					// Randomize how many times the request is sent, for burst volume
					for i := 0; i <= rand.Intn(3); i++ {
						time.Sleep(200 * time.Millisecond)
						event.setDsn(dsn)
						request := NewRequest(event)
						request.send()
					}
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
