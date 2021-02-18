package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"github.com/getsentry/sentry-go"
	"google.golang.org/api/iterator"
)

type DemoAutomation struct{}

// Get events from both Sentry and GCS
func (d *DemoAutomation) getEvents() []Event {
	var events []Event
	events1 := d.getEventsFromSentry()
	events2 := d.getEventsFromGCS()
	events = append(events, events1...)
	events = append(events, events2...)
	return events
}

// Download the events from Sentry. You may have to be a team member on the org you're downloading events from (SENTRY_AUTH_TOKEN)
func (d *DemoAutomation) getEventsFromSentry() []Event {
	var events []Event

	discoverAPI := DiscoverAPI{}
	eventsAPI := EventsAPI{}

	for _, org := range config.Sources {
		eventMetadata := discoverAPI.latestEventMetadata(org, *n)
		_events := eventsAPI.getEvents(org, eventMetadata)
		events = append(events, _events...)
	}
	fmt.Printf("\n> EVENTS from API: %v \n", len(events))
	return events
}

// Gets events from Google Cloud Storage
func (d *DemoAutomation) getEventsFromGCS() []Event {
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
	bucketName := config.Bucket
	bucketHandle := client.Bucket(bucketName)

	var fileNames []string

	query := &storage.Query{Prefix: *filePrefix}
	it := bucketHandle.Objects(ctx, query)
	for {
		obj, err := it.Next()
		if err == iterator.Done {
			sentry.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTag("files", fmt.Sprint(len(fileNames)))
			})
			sentry.CaptureMessage(fmt.Sprintf("finished retrieving files"))
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

		event.setPlatform()
		event.undertake()
		events = append(events, event)
		events = removeMechanism(events)
	}
	return events
}

func removeMechanism(_events []Event) []Event {
	for _, event := range _events {
		if event.Kind == ERROR || event.Kind == DEFAULT {
			exception := event.Error.Exception
			values := exception["values"]
			if values != nil {
				for _, value := range values.([]interface{}) {
					mechanism := value.(map[string]interface{})["mechanism"]
					if mechanism != nil {
						mechanismType := mechanism.(map[string]interface{})["type"]
						if mechanismType == "minidump" {
							delete(value.(map[string]interface{}), "mechanism")
						}
					}
				}
			}
		}
	}
	return _events
}

func printObj(obj *storage.ObjectAttrs) {
	fmt.Printf("filename: /%v/%v \n", obj.Bucket, obj.Name) // .ContentType .Owner .Size
}
