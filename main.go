package main

import (
	"context"
	// "encoding/json"
	"flag"

	// "io/ioutil"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	// "google.golang.org/api/iterator"
)

var httpClient = &http.Client{} // TODO

var (
	all    *bool
	id     *string
	ignore *bool
	db     *string // or similar could be for dynamically handling the 'prefix' for Bucket
	js     *string
	py     *string
	// dsn         DSN
	SENTRY_URL  string
	projectDSNs map[string]*DSN
	traceIds    []string

	// TESTING...
	client       *storage.Client
	ctx          context.Context
	bucketHandle *storage.BucketHandle
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	all = flag.Bool("all", false, "send all events. default is send latest event")
	// id = flag.String("id", "", "id of event in sqlite database") // 08/27 non-functional today
	ignore = flag.Bool("i", false, "ignore sending the event to Sentry.io")
	js = flag.String("js", "", "javascript DSN")
	py = flag.String("py", "", "python DSN")
	flag.Parse()

	projectDSNs = make(map[string]*DSN)
	projectDSNs["javascript"] = parseDSN(os.Getenv("DSN_JAVASCRIPT_SAAS"))
	if *js != "" {
		projectDSNs["javascript"] = parseDSN(*js)
	}
	projectDSNs["python"] = parseDSN(os.Getenv("DSN_PYTHON_SAAS"))
	if *py != "" {
		projectDSNs["python"] = parseDSN(*py)
	}
}

func main() {

	demoAutomation := DemoAutomation{}

	events := demoAutomation.getEvents() // configureDsns() too ?

	for _, event := range events {
		if event.Kind == "error" {
			event.Error.eventId()
			event.Error.release()
			event.Error.user()
			event.Error.timestamp()
		}
		if event.Kind == "transaction" {
			event.Transaction.eventId()
			event.Transaction.release()
			event.Transaction.user()
			event.Transaction.timestamps()
		}
	}

	getTraceIds(events)
	updateTraceIds(events)

	requests := Requests{events}
	requests.send()

	return
}

// demoAutomation.init()

// // 1 ctx, client
// ctx := context.Background()
// client, err := storage.NewClient(ctx)
// if err != nil {
// 	log.Fatalln("storage.NewClient:", err)
// }
// defer client.Close()
// ctx, cancel := context.WithTimeout(ctx, time.Second*50)
// defer cancel()

// // 2 bucket handle
// bucketName := os.Getenv("BUCKET")
// bucketHandle := client.Bucket(bucketName)

// lists the contents of a bucket in Google Cloud Storage.
// demoAutomation.getFileNames()

// var fileNames []string
// query := &storage.Query{Prefix: "eventtest"}
// it := demoAutomation.BucketHandle.Objects(demoAutomation.Ctx, query)
// // it := bucketHandle.Objects(ctx, query)
// for {
// 	obj, err := it.Next()
// 	if err == iterator.Done {
// 		break
// 	}
// 	if err != nil {
// 		log.Fatalln("listBucket: unable to list bucket", err)
// 	}
// 	fileNames = append(fileNames, obj.Name)
// 	printObj(obj)
// }

// Read each file's content
// var events []Event
// for _, fileName := range fileNames {
// 	rc, err := demoAutomation.BucketHandle.Object(fileName).NewReader(demoAutomation.Ctx)
// 	// rc, err := bucketHandle.Object(fileName).NewReader(ctx)
// 	if err != nil {
// 		log.Fatalln("NewReader:", err)
// 	}
// 	byteValue, _ := ioutil.ReadAll(rc)

// 	// The Event's UnmarshalJSON overriden in Event.go
// 	var event Event
// 	if err := json.Unmarshal(byteValue, &event); err != nil {
// 		panic(err)
// 	}
// 	events = append(events, event)
// }

// TODO double check it's object was updated reference `fmt.Println("\n> timestamp AFTER", event.Error.Timestamp)`
