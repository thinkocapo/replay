package main

import (
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"io/ioutil"
	"log"
	"net/http"
)

type EventsAPI struct{}

func (e EventsAPI) getEvents(org string, eventMetadata []EventMetadata) []Event {
	var events []Event

	for _, e := range eventMetadata {
		if e.Project == config.Skip {
			continue
		}

		endpoint := "https://sentry.io/api/0/projects/" + org + "/" + e.Project + "/events/" + e.Id + "/json/"

		request, _ := http.NewRequest("GET", endpoint, nil)
		request.Header.Set("content-type", "application/json")
		request.Header.Set("Authorization", fmt.Sprint("Bearer ", config.SentryAuthToken))

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
		event.undertake()
		events = append(events, event)
	}
	events = sanitizeOrg(events)
	events = fingerprintCheck(events)
	fmt.Printf("> %v Events length %v\n", org, len(events))
	return events
}

func fingerprintCheck(_events []Event) []Event {
	for _, event := range _events {
		if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == JAVASCRIPT {
			metadata := event.Error.Metadata
			// stack.abs_path is different on each (due to static.js/testing being used), thereby creating too many unique issues
			if metadata["type"] == "AssertionError" && metadata["value"] == "expected 'Error' to equal 'TypeError'" {
				event.Error.Fingerprint = []string{"assertion-error-expected"}
			}
		}
	}
	return _events
}

func hasOrgTag(event Event) bool {
	var tags [][]string
	if event.Kind == ERROR || event.Kind == DEFAULT {
		tags = event.Error.Tags
	}
	if event.Kind == TRANSACTION {
		tags = event.Transaction.Tags
	}

	for _, tag := range tags {
		if tag[0] == "organization" {
			fmt.Println("\n> has organization tag")
			return true
		}
	}
	return false
}

func sanitizeOrg(_events []Event) []Event {
	var events []Event
	for _, event := range _events {
		if hasOrgTag(event) == false {
			events = append(events, event)
		}
	}
	return events
}
