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

type Discover struct {
	Data []map[string]interface{} `json:"data"`
}

func (d Discover) latestEventList() string {
	org := os.Getenv("ORG")
	n := 15

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
	// 1 eventMinis := discover.parseEventMinis()
	// eventList := d.Data
	fmt.Println("* * * ** * * len(d.Data) * * * * *", len(d.Data))

	// ...

	return "testing"
}
