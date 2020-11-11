package main

import (
	"context"
	"fmt"
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
	query := &storage.Query{Prefix: "event"}
	it := bucketHandle.Objects(ctx, query)
	for {
		obj, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal("listBucket: unable to list bucket %q: %v", bucketHandle, err)
		}
		printObj(obj)
		// fmt.Println(">>>>> obj", obj)
		// d.dumpStats(obj)
	}

	// read each file content
	// file := database
	// fmt.Println("DATASET file", file)
	// rc, err := client.Bucket(bucket).Object(file).NewReader(ctx)
	// if err != nil {
	// 	log.Fatalln("NewReader:", err)
	// }
	// byteValue, _ := ioutil.ReadAll(rc) // jsonFile
	// // defer jsonFile.Close()
	// events := make([]EventJson, 0)
	// if err := json.Unmarshal(byteValue, &events); err != nil {
	// 	panic(err)
	// }

	return "read those jsons"
}

func printObj(obj *storage.ObjectAttrs) {
	fmt.Printf("(filename: /%v/%v, \n", obj.Bucket, obj.Name)
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
