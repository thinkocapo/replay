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
		if e.Project == os.Getenv("SKIP") {
			fmt.Println("skipping...", e.Project)
			continue
		}

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
		event.undertake()
		// TODO could sanitize/flag it here, and then not append it. organization.slug, plan.tier
		events = append(events, event)
	}
	events = sanitize(events)
	fmt.Printf("> %v Events length %v\n", org, len(events))
	return events
}

func sanitize(_events []Event) []Event {
	var events []Event

	for _, event := range _events {
		if hasOrgTag(event) == false {
			events = append(events, event)
		} else {
			fmt.Println("\n > has org!!")
		}
	}
	return events
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
			fmt.Println("\n> has org tag!")
			return true
		}
	}
	return false
}
