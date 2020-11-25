package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/getsentry/sentry-go"
)

type DiscoverAPI struct {
	Data     []EventMetadata
	endpoint string
}

type EventMetadata struct {
	Id       string
	Project  string
	Platform string
}

// Events from last 24HrPeriod events for selected Projects
// Returns event metadata (e.g. Id, Project) but not the entire Event itself, which gets queried separately.
func (d DiscoverAPI) latestEventMetadata(org string, n int) []EventMetadata {
	// org := os.Getenv("ORG")

	query := "&query=platform.name%3Ajavascript+OR+platform.name%3Apython"

	// with 0 project names specified
	// endpoint := fmt.Sprintf("https://sentry.io/api/0/organizations/%v/eventsv2/?statsPeriod=24h&field=event.type&field=project&field=platform&per_page=%v&query=%v", org, strconv.Itoa(n), query)

	// with 2 project names specified da-flask da-react
	endpoint := fmt.Sprintf("https://sentry.io/api/0/organizations/%v/eventsv2/?statsPeriod=24h&project=5422148&project=5427415&field=event.type&field=project&field=platform&per_page=%v&query=%v", org, strconv.Itoa(n), query)

	// with 1 project name specified da-react
	// endpoint := fmt.Sprintf("https://sentry.io/api/0/organizations/%v/eventsv2/?statsPeriod=24h&project=5427415&field=event.type&field=project&field=platform&per_page=%v&query=%v", org, strconv.Itoa(n), query)

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

	json.Unmarshal(body, &d)

	fmt.Println("> Data []EventMetadata  length:", len(d.Data))

	// TODO FOR TESTING the org filtering
	for _, e := range d.Data {
		fmt.Println("> Project", e.Project)
	}
	return d.Data
}

// Consider
// func (d DiscoverAPI) setPlatform(platform string) DiscoverAPI {
// 	// TODO builder
// 	// d.endpoint := query
// 	return d
// }

// Consider
// func (d DiscoverAPI) execute() {
// //
// }
