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
	// OG
	// Data []map[string]interface{} `json:"data"`

	Data     []EventMetadata
	endpoint string

	// or
	// query string
	// params string
}

type EventMetadata struct {
	Id       string
	Project  string
	Platform string // `json:"platform.name"`
}

// Events from last 24HrPeriod events for selected Projects
// Returns event metadata (e.g. Id, Project) but not the entire Event itself, which gets queried separately.
func (d DiscoverAPI) latestEventMetadata(n int) []EventMetadata {
	org := os.Getenv("ORG")

	// DEPRECATE don't need all these extra columns
	// endpoint := "https://sentry.io/api/0/organizations/" + org + "/eventsv2/?statsPeriod=24h&project=5422148&project=5427415&field=title&field=event.type&field=project&field=user.display&field=timestamp&sort=-timestamp&per_page=" + strconv.Itoa(n) + "&query="

	// OG, Still has Project IDs
	// endpoint := "https://sentry.io/api/0/organizations/" + org + "/eventsv2/?statsPeriod=24h&project=5422148&project=5427415&field=event.type&field=project&field=platform&per_page=" + strconv.Itoa(n) + "&query="

	// ATTEMPT no project ids
	query := "&query=platform.name%3Ajavascript+OR+platform.name%3Apython"
	endpoint := fmt.Sprintf("https://sentry.io/api/0/organizations/%v/eventsv2/?statsPeriod=24h&field=event.type&field=project&field=platform&per_page=%v&query=%v", org, strconv.Itoa(n), query)
	fmt.Println("> > ENDPOINT", endpoint)

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
	return d.Data

	// TODO for .execute()
	// return d
}

func (d DiscoverAPI) setPlatform(platform string) DiscoverAPI {
	// TODO builder
	// d.endpoint := query
	return d
}

// DEPRECATE - Select Platform
// func (d DiscoverAPI) platform(platform string) DiscoverAPI {
// 	fmt.Print("> SELECTIT len(Data)", len(d.Data))
// 	for _, eventMetadata := range d.Data {
// 		if eventMetadata.Platform != platform {
// 			fmt.Println("> > platform was", eventMetadata.Platform)
// 		} else {
// 			fmt.Println("> > platform is", eventMetadata.Platform)
// 		}
// 	}
// 	return d
// }

func (d DiscoverAPI) get() []EventMetadata {
	return d.Data
}

// }
// idea
// d.EventMetadatas = eventMetadatas
