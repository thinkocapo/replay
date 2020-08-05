
// // Sample storage-quickstart creates a Google Cloud Storage bucket.
// package main

// import (
//         "context"
//         "fmt"
//         "log"
//         "time"

//         "cloud.google.com/go/storage"
// )

// func main() {
//         ctx := context.Background()

//         // Sets your Google Cloud Platform project ID.
//         projectID := "sales-engineering-sf"

//         // Creates a client.
//         client, err := storage.NewClient(ctx)
//         if err != nil {
//                 log.Fatalf("Failed to create client: %v", err)
//         }

//         // Sets the name for the new bucket.
//         bucketName := "undertakerevents"

//         // Creates a Bucket instance.
//         bucket := client.Bucket(bucketName)

//         // Creates the new bucket.
//         ctx, cancel := context.WithTimeout(ctx, time.Second*10)
//         defer cancel()
//         if err := bucket.Create(ctx, projectID, nil); err != nil {
//                 log.Fatalf("Failed to create bucket: %v", err)
//         }

//         fmt.Printf("Bucket %v created.\n", bucketName)
// }