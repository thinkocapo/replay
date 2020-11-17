package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/api/iterator"
)

var httpClient = &http.Client{}

var (
	all         *bool
	id          *string
	ignore      *bool
	database    string
	db          *string
	js          *string
	py          *string
	dsn         DSN
	SENTRY_URL  string
	exists      bool
	projectDSNs map[string]*DSN
	traceIds    []string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	all = flag.Bool("all", false, "send all events. default is send latest event")
	// id = flag.String("id", "", "id of event in sqlite database") // 08/27 non-functional today
	ignore = flag.Bool("i", false, "ignore sending the event to Sentry.io")
	// db = flag.String("db", "", "database.json")
	js = flag.String("js", "", "javascript DSN")
	py = flag.String("py", "", "python DSN")
	flag.Parse()

	// TODO demoAutomation.Dsns.configure() or demoAutomation.configureDsns()
	projectDSNs = make(map[string]*DSN)
	projectDSNs["javascript"] = parseDSN(os.Getenv("DSN_JAVASCRIPT_SAAS"))
	if *js != "" {
		projectDSNs["javascript"] = parseDSN(*js)
	}
	projectDSNs["python"] = parseDSN(os.Getenv("DSN_PYTHON_SAAS"))
	if *py != "" {
		projectDSNs["python"] = parseDSN(*py)
	}

	// if *db == "" {
	// 	database = os.Getenv("JSON")
	// } else {
	// 	database = *db
	// }
}

type DemoAutomation struct {
	client                *storage.Client
	ctx                   context.Context
	bucketHandle          *storage.BucketHandle // `client.Bucket(bucketName)` for setting this
	bucketHandleFileNames []string              // `query := &storage.Query{Prefix: "eventtest"}` for setting this
	// TODO consider `events []EventJson` ?
	// TODO consider setDsns...
}

// TODO METHODS
// Constructor/Init for DemoAutomation
// then can call demoAutomation.client() demoAutomation.bucketHandle.Objects()

func main() {
	// TODO da | demoAutomation

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

	// TODO events.go could manage reading from storage. or like:
	/*
		storageClient := StorageClient(os.Getenv("BUCKET")) <-- is the init
		//or
		demoAutomation.init(os.Getenv("BUCKET"))
		demoAutomation.query("event") .prefixQuery("event") .queryBucket .bucketQuery() .bucketSet()
		demoAutomation.listBucketContents() .getBucket()
		events := demoAutomation.getFiles() .bucketFiles()
	*/

	// Read each file's content
	var events []EventJson
	for _, fileName := range fileNames {
		rc, err := bucketHandle.Object(fileName).NewReader(ctx)
		if err != nil {
			log.Fatalln("NewReader:", err)
		}
		byteValue, _ := ioutil.ReadAll(rc)
		var event EventJson
		// The EventJson's UnmarshalJSON overriden in event-to-sentry.go (soon EventJson.go)
		if err := json.Unmarshal(byteValue, &event); err != nil {
			panic(err)
		}
		events = append(events, event)
	}

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

	// TODO double check it's object was updated reference `fmt.Println("\n> timestamp AFTER", event.Error.Timestamp)`
	requests := Requests{events}
	requests.send()

	return
}

func printObj(obj *storage.ObjectAttrs) {
	fmt.Printf("filename: /%v/%v \n", obj.Bucket, obj.Name)
	// fmt.Printf("ContentType: %q, ", obj.ContentType)
	// fmt.Printf("Owner: %v, ", obj.Owner)
	// fmt.Printf("Size: %v, ", obj.Size)
}
