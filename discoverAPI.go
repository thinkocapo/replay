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
// n of 200 may not work
func (d DiscoverAPI) latestEventMetadata(org string, n int) []EventMetadata {
	fmt.Printf("\n> ORG %v\n", org)

	// query := makeQuery([]string{JAVASCRIPT, PYTHON, JAVA, RUBY, GO, NODE, PHP, CSHARP, DART, ELIXIR, PERL, RUST, COCOA, ANDROID})
	query := makeQuery(platforms)
	endpoint := fmt.Sprintf("https://sentry.io/api/0/organizations/%v/eventsv2/?statsPeriod=24h&field=event.type&field=project&field=platform&per_page=%v&query=%v", org, strconv.Itoa(n), query)

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
	// slice the final "+OR+" off
	return result[:len(result)-4]
}

// Consider chaining
// func (d DiscoverAPI) execute() {
// //
// }
