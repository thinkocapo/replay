package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/getsentry/sentry-go"
	"google.golang.org/api/iterator"
)

type DemoAutomation struct{}

const JAVASCRIPT = "javascript"
const PYTHON = "python"

// download the events from Sentry
func (d *DemoAutomation) downloadEvents() []Event {
	org := os.Getenv("ORG")
	var eventIds []string
	var events []Event
	n := 10

	// Call Sentry w/ 24HrPeriod events with Projects selected
	// TODO could get pg 2 after
	endpoint := fmt.Sprint("https://sentry.io/api/0/organizations/", org, "/eventsv2/?statsPeriod=24h&project=5260888&project=1428657&field=title&field=event.type&field=project&field=user.display&field=timestamp&sort=-timestamp&per_page=", n, "&query=")

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

	var discover Discover
	json.Unmarshal(body, &discover)
	eventMinis := discover.Data
	for _, e := range eventMinis {
		// eventId := event.(map[string]interface{})["eventId"]
		eventIds = append(eventIds, e["id"].(string))
	}
	fmt.Println("\n> > > > > > > > # eventIds > > > > > > > >", len(eventIds))

	for _, id := range eventIds {
		// 	// TODO Call JSON URL for each
		endpoint2 := fmt.Sprint("https://sentry.io/api/0/projects/", org, "/will-frontend-react/events/", id, "/json/")
		request2, _ := http.NewRequest("GET", endpoint2, nil)

		request2.Header.Set("content-type", "application/json")
		request2.Header.Set("Authorization", fmt.Sprint("Bearer ", os.Getenv("SENTRY_AUTH_TOKEN")))

		var httpClient = &http.Client{}
		response2, requestErr2 := httpClient.Do(request2)
		if requestErr2 != nil {
			sentry.CaptureException(requestErr2)
			log.Fatal(requestErr)
		}
		body2, errResponse2 := ioutil.ReadAll(response2.Body)
		if errResponse2 != nil {
			sentry.CaptureException(errResponse2)
			log.Fatal(errResponse2)
		}

		var event Event
		// TODO - may need to eliminate first 2 lines which are comments
		json.Unmarshal(body2, &event)
		event.setDsn()
		events = append(events, event)

		// 	byteValue, _ := ioutil.ReadAll(somethingThatReadSentry)
		// 	var event Event
		// 	if err := json.Unmarshal(byteValue, &event); err != nil {
		// 		sentry.CaptureException(err)
		// 		panic(err)
		// 	}
		// 	event.setDsn()
		// 	events = append(events, event)
	}

	return events
}

// get the events from GCS
func (d *DemoAutomation) getEvents(filePrefix string) []Event {
	// Initialize/Connect the Client
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		sentry.CaptureException(err)
		log.Fatalln("storage.NewClient:", err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Get the bucket and its file names
	bucketName := os.Getenv("BUCKET")
	bucketHandle := client.Bucket(bucketName)

	var fileNames []string

	query := &storage.Query{Prefix: filePrefix}
	it := bucketHandle.Objects(ctx, query)
	for {
		obj, err := it.Next()
		if err == iterator.Done {
			sentry.CaptureMessage(fmt.Sprintf("finished retrieving %v file names", len(fileNames)))
			break
		}
		if err != nil {
			sentry.CaptureException(err)
			log.Fatalln("listBucket: unable to list bucket", err)
		}
		fileNames = append(fileNames, obj.Name)
		print(obj)
	}

	// Get the files
	var events []Event
	for _, fileName := range fileNames {
		rc, err := bucketHandle.Object(fileName).NewReader(ctx)
		if err != nil {
			sentry.CaptureException(err)
			log.Fatalln("NewReader:", err)
		}
		byteValue, _ := ioutil.ReadAll(rc)

		// Dev Note - The Event's UnmarshalJSON method is overriden in Event.go
		var event Event
		if err := json.Unmarshal(byteValue, &event); err != nil {
			sentry.CaptureException(err)
			panic(err)
		}
		event.setDsn()
		events = append(events, event)
	}
	return events
}

func print(obj *storage.ObjectAttrs) {
	fmt.Printf("filename: /%v/%v \n", obj.Bucket, obj.Name) // .ContentType .Owner .Size
}
