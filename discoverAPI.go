package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
// n of 200 may not work
func (d DiscoverAPI) latestEventMetadata(org string, n int) []EventMetadata {
	query := makeQuery(platforms)
	endpoint := fmt.Sprintf("https://sentry.io/api/0/organizations/%v/eventsv2/?statsPeriod=24h&field=event.type&field=project&field=platform&per_page=%v&query=%v", org, strconv.Itoa(n), query)

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

	json.Unmarshal(body, &d)
	fmt.Printf("> %v Discover.Data length: %v\n", org, len(d.Data))

	for _, e := range d.Data {
		fmt.Printf("> %v %v\n", org, e.Project)
	}
	return d.Data
}

func makeQuery(supportedPlatforms []string) string {
	var result string
	for _, platform := range supportedPlatforms {
		result += "platform.name%3A" + platform + "+OR+"
	}
	// slices the final "+OR+" off
	return result[:len(result)-4]
}

// EVAL chaining
// func (d DiscoverAPI) execute() {
//
// }
