package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
)

type EventsAPI struct{}

func (e EventsAPI) getEvents(eventMetadata []EventMetadata) []Event {
	org := os.Getenv("ORG")
	var events []Event

	for _, e := range eventMetadata {
		endpoint2 := fmt.Sprint("https://sentry.io/api/0/projects/", org, "/", e.Project, "/events/", e.Id, "/json/")

		request2, _ := http.NewRequest("GET", endpoint2, nil)
		request2.Header.Set("content-type", "application/json")
		request2.Header.Set("Authorization", fmt.Sprint("Bearer ", os.Getenv("SENTRY_AUTH_TOKEN")))

		var httpClient = &http.Client{}
		response2, requestErr2 := httpClient.Do(request2)
		if requestErr2 != nil {
			sentry.CaptureException(requestErr2)
			log.Fatal(requestErr2)
		}
		body2, errResponse2 := ioutil.ReadAll(response2.Body)
		if errResponse2 != nil {
			// TODO - could already be a bad response, if wrong project name requested
			sentry.CaptureException(errResponse2)
			log.Fatal(errResponse2)
		}

		var event Event
		// json.Unmarshal(body2, &event)
		if err2 := json.Unmarshal(body2, &event); err2 != nil {
			sentry.CaptureException(err2)
			panic(err2)
		}
		event.setDsn()
		events = append(events, event)
	}
	return events
}
