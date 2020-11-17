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

// wanted...
// Constructor/Init for DemoAutomation
// then can call demoAutomation.client() demoAutomation.bucketHandle.Objects()
// TODO events.go could manage reading from storage. or like:
/*
	storageClient := StorageClient(os.Getenv("BUCKET")) <-- is the init
	//or
	demoAutomation.init(os.Getenv("BUCKET"))
	demoAutomation.query("event") .prefixQuery("event") .queryBucket .bucketQuery() .bucketSet()
	demoAutomation.listBucketContents() .getBucket()
	events := demoAutomation.getFiles() .bucketFiles()
*/

// TODO could take care of DSN's still?
type DemoAutomation struct {
	Client       *storage.Client
	Ctx          context.Context
	BucketHandle *storage.BucketHandle // `client.Bucket(bucketName)` for setting this
	fileNames    []string              // `query := &storage.Query{Prefix: "eventtest"}` for setting this
	// TODO consider `events []Event` ?
	// TODO consider setDsns... for projectDSNs
}

// TODO Constructor that configures DSN's

func (d *DemoAutomation) getEvents() []Event {
	// 1 Initialize/Connect the Client
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalln("storage.NewClient:", err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// d.Ctx = ctx
	// d.Client = client

	// 2
	bucketName := os.Getenv("BUCKET")
	bucketHandle := client.Bucket(bucketName)

	var fileNames []string
	query := &storage.Query{Prefix: "eventtest"}
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

	var events []Event
	for _, fileName := range fileNames {
		rc, err := bucketHandle.Object(fileName).NewReader(ctx)
		if err != nil {
			log.Fatalln("NewReader:", err)
		}
		byteValue, _ := ioutil.ReadAll(rc)

		// Dev Note - The Event's UnmarshalJSON method is overriden in Event.go
		var event Event
		if err := json.Unmarshal(byteValue, &event); err != nil {
			panic(err)
		}
		events = append(events, event)
	}
	return events
}

func printObj(obj *storage.ObjectAttrs) {
	fmt.Printf("filename: /%v/%v \n", obj.Bucket, obj.Name)
	// fmt.Printf("ContentType: %q, ", obj.ContentType)
	// fmt.Printf("Owner: %v, ", obj.Owner)
	// fmt.Printf("Size: %v, ", obj.Size)
}

// func NewDemoAutomation() *DemoAutomation {
// 	fmt.Print("111111")
// 	// 1 Initialize/Connect the Client
// 	ctx := context.Background()
// 	client, err := storage.NewClient(ctx)
// 	if err != nil {
// 		log.Fatalln("storage.NewClient:", err)
// 	}
// 	defer client.Close()
// 	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
// 	defer cancel()

// 	// d.Ctx = ctx
// 	// d.Client = client

// 	// 2
// 	bucketName := os.Getenv("BUCKET")
// 	bucketHandle := client.Bucket(bucketName)
// 	// d.BucketHandle = bucketHandle
// 	fmt.Print("2222")

// 	d := &DemoAutomation{
// 		Ctx:    ctx,
// 		Client: client,
// 		// bucket:     client.Bucket(bucket),
// 		BucketHandle: bucketHandle,
// 	}
// 	return d
// }

// TODO consider using natural Constructor instead
// func (d *DemoAutomation) init() {
// 	fmt.Print("111111")
// 	// 1 Initialize/Connect the Client
// 	ctx := context.Background()
// 	client, err := storage.NewClient(ctx)
// 	if err != nil {
// 		log.Fatalln("storage.NewClient:", err)
// 	}
// 	defer client.Close()
// 	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
// 	defer cancel()

// 	d.Ctx = ctx
// 	d.Client = client

// 	// 2
// 	bucketName := os.Getenv("BUCKET")
// 	bucketHandle := client.Bucket(bucketName)
// 	d.BucketHandle = bucketHandle
// 	fmt.Print("2222")
// }

// func (d *DemoAutomation) getFileNames() {
// 	var fileNames []string
// 	query := &storage.Query{Prefix: "eventtest"}
// 	it := d.BucketHandle.Objects(d.Ctx, query)
// 	// it := bucketHandle.Objects(ctx, query)
// 	for {
// 		obj, err := it.Next()
// 		if err == iterator.Done {
// 			break
// 		}
// 		if err != nil {
// 			log.Fatalln("listBucket: unable to list bucket", err)
// 		}
// 		fileNames = append(fileNames, obj.Name)
// 		printObj(obj)
// 	}
// }
