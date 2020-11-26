package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/getsentry/sentry-go"
	"google.golang.org/api/iterator"
)

type DemoAutomation struct{}

const JAVASCRIPT = "javascript"
const PYTHON = "python"

// Download the events from Sentry
func (d *DemoAutomation) getEventsFromSentry() []Event {
	var events []Event
	// sources := readSources()
	// fmt.Printf("> SOURCES length: %v \n", len(sources))

	discoverAPI := DiscoverAPI{}
	eventsAPI := EventsAPI{}

	for _, org := range config.Sources {
		// TODO, switch 'n' to ./cli flag
		eventMetadata := discoverAPI.latestEventMetadata(org, *n)
		_events := eventsAPI.getEvents(org, eventMetadata)

		fmt.Printf("> %v Events length %v\n", org, len(_events))
		events = append(events, _events...)
	}
	fmt.Printf("> FINAL EVENTS length: %v \n", len(events))
	return events
}

// Get the events from Google Cloud Storage
// ./bin/main -i eventtest
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
		printObj(obj)
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

		// TODO may be broken, now that setDsn changed.
		event.setDsnGCS()
		events = append(events, event)
	}
	return events
}

func printObj(obj *storage.ObjectAttrs) {
	fmt.Printf("filename: /%v/%v \n", obj.Bucket, obj.Name) // .ContentType .Owner .Size
}
