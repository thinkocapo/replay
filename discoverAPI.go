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

type DiscoverAPI struct {
	Data []map[string]interface{} `json:"data"`
}

type EventMetadata struct {
	Id      string
	Project string
}

// Events from last 24HrPeriod events for selected Projects
// Returns event metadata (e.g. Id, Project) but not the entire Event itself, which gets queried separately.
func (d DiscoverAPI) latestEventMetadata(n int) []EventMetadata {
	org := os.Getenv("ORG")

	endpoint := fmt.Sprint("https://sentry.io/api/0/organizations/", org, "/eventsv2/?statsPeriod=24h&project=5422148&project=5427415&field=title&field=event.type&field=project&field=user.display&field=timestamp&sort=-timestamp&per_page=", n, "&query=")
	request, _ := http.NewRequest("GET", endpoint, nil)
	request.Header.Set("content-type", "application/json")
	request.Header.Set("Authorization", fmt.Sprint("Bearer ", os.Getenv("SENTRY_AUTH_TOKEN")))

	var httpClient = &http.Client{}
	response, requestErr := httpClient.Do(request)
	if requestErr != nil {
		sentry.CaptureException(requestErr)
		log.Fatal(requestErr)
	}
	body, errResponse := ioutil.ReadAll(response.Body)
	if errResponse != nil {
		sentry.CaptureException(errResponse)
		log.Fatal(errResponse)
	}

	json.Unmarshal(body, &d)
	eventMetadata := d.Data

	var eventMetadatas []EventMetadata
	for _, e := range eventMetadata {
		eventMetadata := EventMetadata{e["id"].(string), e["project"].(string)}
		eventMetadatas = append(eventMetadatas, eventMetadata)
	}
	fmt.Println("> eventMetadata length:", len(eventMetadata))
	return eventMetadatas
}
