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
)

func readJsons() string {
	fmt.Println("This is a readJsons test...")

	bucket := os.Getenv("BUCKET")

	// Initialize/Connect the Client
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalln("storage.NewClient:", err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// TODO
	// 1 List all files for bucket...

	// 2. iterate through each and add to global...

	file := database
	fmt.Println("DATASET file", file)
	rc, err := client.Bucket(bucket).Object(file).NewReader(ctx)
	if err != nil {
		log.Fatalln("NewReader:", err)
	}
	byteValue, _ := ioutil.ReadAll(rc) // jsonFile
	// defer jsonFile.Close()
	events := make([]EventJson, 0)
	if err := json.Unmarshal(byteValue, &events); err != nil {
		panic(err)
	}

	return "read those jsons"
}
