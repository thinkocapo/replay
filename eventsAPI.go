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

type EventsAPI struct {
	// events []Event
}

func (e EventsAPI) getEvents(org string, eventMetadata []EventMetadata) []Event {
	var events []Event

	for _, e := range eventMetadata {
		endpoint := "https://sentry.io/api/0/projects/" + org + "/" + e.Project + "/events/" + e.Id + "/json/"

		request, _ := http.NewRequest("GET", endpoint, nil)
		request.Header.Set("content-type", "application/json")
		request.Header.Set("Authorization", fmt.Sprint("Bearer ", os.Getenv("SENTRY_AUTH_TOKEN")))

		var httpClient = &http.Client{}
		response, err := httpClient.Do(request)
		if err != nil {
			sentry.CaptureException(err)
			log.Fatal(err)
		}
		body, errResponse := ioutil.ReadAll(response.Body)
		if errResponse != nil {
			sentry.CaptureException(errResponse)
			log.Fatal(errResponse)
		}

		var event Event
		if errUnmarshal := json.Unmarshal(body, &event); errUnmarshal != nil {
			sentry.CaptureException(errUnmarshal)
			panic(errUnmarshal)
		}
		event.setPlatform()
		// TODO could sanitize/flag it here, and then not append it. organization.slug, plan.tier
		events = append(events, event)
	}
	return events
}
