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
func readJsons() string {
	fmt.Println("This is a readJsons test...")

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
	fmt.Println(">>>>> files", len(fileNames))

	// Read each file's content
	var events []EventJson
	for _, fileName := range fileNames {
		fmt.Println("> file", fileName)
		rc, err := bucketHandle.Object(fileName).NewReader(ctx)
		if err != nil {
			log.Fatalln("NewReader:", err)
		}
		byteValue, _ := ioutil.ReadAll(rc) // jsonFile
		// defer jsonFile.Close()
		// event := make([]EventJson, 0)
		var event EventJson
		if err := json.Unmarshal(byteValue, &event); err != nil {
			panic(err)
		}

		events = append(events, event)
	}

	fmt.Println(">>>>> events []EventJson", len(events))
	for _, event := range events {
		// match DSN based on js vs python
		if event.Type == "error" {
			fmt.Println("> error")
			eventError := Error{event.EventId, event.Release, event.User, event.Timestamp}
			eventError.eventId()
			eventError.release()
			eventError.user()
			eventError.setTimestamp()
		}
		if event.Type == "transaction" {
			fmt.Println("> transaction")
			eventTransaction := Transaction{event.EventId, event.Release, event.User, event.Timestamp}
			eventTransaction.eventIds()
			eventTransaction.setReleases()
			eventTransaction.setUsers()
			eventTransaction.setTimestamps()

			eventTransaction.sentAt()
			eventTransaction.removeLengthField()

		}

		fmt.Println("> event.eventId", event.EventId)
	}

	return "read those jsons"
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
