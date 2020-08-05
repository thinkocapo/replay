package main

import (
	"bytes"
	"context"
	
	_ "github.com/mattn/go-sqlite3"

	"fmt"
	"io"
	"io/ioutil"
	"time"

	"encoding/json"

	"cloud.google.com/go/storage"
)


// Social struct which contains a
// list of links
type Event struct {
    Level string `json:"level"`
	Event_id  string `json:"event_id"`
	Timestamp string `json:"timestamp"`
	Server_name string `json:"server_name"`
	Platform string `json:"platform"`

	// Exception interface{} `json:"exception"`
	// Breadcrumbs interface{} `json:"breadcrumbs"`
	// Context interface{} `json:"context"`
	// Modules interface{} `json:"modules"`
	// Extra interface{} `json:"extra"`
	// Sdk interface{} `json:"sdk"`
}

func main () {
	bucket := "undertakerevents"
	object := "events.json"
	// object := "events2.json"
	// object := "tracing-example-multiproject.db" can't unmarshallJSON on this. it's not JSON it's flat-file db sqlite
	// object := "users.json"
	// object := "testarray.json"

	var buf1 bytes.Buffer
	w := io.Writer(&buf1)

	downloadFile(w, bucket, object)
}


func downloadFile(w io.Writer, bucket, object string) ([]byte, error) {
	fmt.Println("\ndownloadFile")
	
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
			fmt.Println("ERROR", err)
			return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
			return nil, fmt.Errorf("Object(%q).NewReader: %v", object, err)
	}

	data, err := ioutil.ReadAll(rc)
	// fmt.Print(data)	
	// fmt.Print(string(data))

	// events := make([]interface{}, 0)
	events := make([]Event, 0)

	// json.Unmarshal(data, &events)
	
	// "cannot unmarshal string into Go value of type main.Event"
	if err := json.Unmarshal(data, &events); err != nil {
		panic(err)
	}
	
	// prints "[{    } {    }]"
	fmt.Println("events", events)

	event := events[0]
	fmt.Println("WORK2", event.Platform)

	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll: %v", err)
	}
	return data, nil
}

func unmarshalJSON(bytes []byte) map[string]interface{} {
	var _interface map[string]interface{}
	if err := json.Unmarshal(bytes, &_interface); err != nil {
		panic(err)
	}
	return _interface
}
