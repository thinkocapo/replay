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
	"google.golang.org/api/iterator"
)

/*
https://cloud.google.com/appengine/docs/standard/go111/googlecloudstorageclient/read-write-to-cloud-storage
https://github.com/GoogleCloudPlatform/golang-samples/blob/8deb2909eadf32523007fd8fe9e8755a12c6d463/docs/appengine/storage/app.go
*/
func readJsons(ignore bool) string {
	bucketName := os.Getenv("BUCKET")
	// Initialize/Connect the Client
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalln("storage.NewClient:", err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	// Prepare bucket handle
	bucketHandle := client.Bucket(bucketName)
	// lists the contents of a bucket in Google Cloud Storage.
	var fileNames []string
	query := &storage.Query{Prefix: "event"}
	it := bucketHandle.Objects(ctx, query)
	for {
		obj, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalln("listBucket: unable to list bucket", err)
		}
		fileNames = append(fileNames, obj.Name)
		printObj(obj)
	}

	// Read each file's content
	var events []EventJson
	for _, fileName := range fileNames {
		rc, err := bucketHandle.Object(fileName).NewReader(ctx)
		if err != nil {
			log.Fatalln("NewReader:", err)
		}
		byteValue, _ := ioutil.ReadAll(rc)
		var event EventJson
		// UnmarshalJSON overriden in error.go
		if err := json.Unmarshal(byteValue, &event); err != nil {
			panic(err)
		}
		events = append(events, event)
	}

	for _, event := range events {
		// TODO match DSN here based on js vs python, call on EventJson?
		if event.Kind == "error" {
			fmt.Println("> error <")
			// fmt.Println("\n> event_id BEFORE", event.Error.EventId)
			// fmt.Println("\n> timestamp BEFORE", event.Error.Timestamp)

			event.Error.eventId()
			event.Error.release()
			event.Error.user()
			event.Error.setTimestamp()

			// fmt.Println("\n> event_id AFTER", event.Error.EventId)
			// fmt.Println("\n> timestamp AFTER", event.Error.Timestamp)
		}
		if event.Kind == "transaction" {
			fmt.Println("> transaction <")

			// TODO 11:23a****** MAY SOLVE IT!!!!
			event.Transaction.eventId()

			// event.Transaction.setReleases()
			// event.Transaction.setUsers()
			// event.Transaction.setTimestamps()

			// event.Transaction.sentAt()
			// event.Transaction.removeLengthField()

			// requests = append(requests, Request{
			// 	EventJson:     event,
			// 	storeEndpoint: dsnToStoreEndpoint(projectDSNs, event.Error.Platform),
			// })
		}

		// TODO can run once here `requests = append(requests, Request{` instead of inside Error as well as Transaction if-then block
	}

	// i.e. put the 'Request' as an embedded struct type on the EventJson ;)??

	// TODO double check it's object was updated reference `fmt.Println("\n> timestamp AFTER", event.Error.Timestamp)`
	requests := Requests{events}
	requests.send()

	return "\n DONE \n"
	// sendRequests(requests, ignore) // deprecate
}

func printObj(obj *storage.ObjectAttrs) {
	fmt.Printf("filename: /%v/%v \n", obj.Bucket, obj.Name)
	// fmt.Printf("ContentType: %q, ", obj.ContentType)
	// fmt.Printf("ACL: %#v, ", obj.ACL)
	// fmt.Printf("Owner: %v, ", obj.Owner)
	// fmt.Printf("ContentEncoding: %q, ", obj.ContentEncoding)
	// fmt.Printf("Size: %v, ", obj.Size)
	// fmt.Printf("MD5: %q, ", obj.MD5)
	// fmt.Printf("CRC32C: %q, ", obj.CRC32C)
	// fmt.Printf("Metadata: %#v, ", obj.Metadata)
	// fmt.Printf("MediaLink: %q, ", obj.MediaLink)
	// fmt.Printf("StorageClass: %q, ", obj.StorageClass)
}
