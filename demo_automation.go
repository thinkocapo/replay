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
func (d *DemoAutomation) getEventsFromSentry() []Event {
	var events []Event
	org := os.Getenv("ORG")

	// Call Sentry w/ 24HrPeriod events with Projects selected
	// TODO could get pg 2 after
	discover := Discover{}
	latestEventList := discover.latestEventList()
	fmt.Println("latestEventList length:", len(latestEventList))

	for _, e := range latestEventList {
		endpoint2 := fmt.Sprint("https://sentry.io/api/0/projects/", org, "/", e.Project, "/events/", e.Id, "/json/")

		request2, _ := http.NewRequest("GET", endpoint2, nil)
		request2.Header.Set("content-type", "application/json")
		request2.Header.Set("Authorization", fmt.Sprint("Bearer ", os.Getenv("SENTRY_AUTH_TOKEN")))

		var httpClient = &http.Client{}
		response2, requestErr2 := httpClient.Do(request2)
		if requestErr2 != nil {
			sentry.CaptureException(requestErr2)
			log.Fatal(requestErr2)
		}
		body2, errResponse2 := ioutil.ReadAll(response2.Body)
		if errResponse2 != nil {
			// TODO - could already be a bad response, if wrong project name requested
			sentry.CaptureException(errResponse2)
			log.Fatal(errResponse2)
		}

		var event Event
		// json.Unmarshal(body2, &event)
		if err2 := json.Unmarshal(body2, &event); err2 != nil {
			fmt.Println("****err2", err2)
			sentry.CaptureException(err2)
			panic(err2)
		}
		event.setDsn()
		events = append(events, event)
	}

	return events
}

func (d *DemoAutomation) getEventsFromGCS(filePrefix string) []Event {
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
