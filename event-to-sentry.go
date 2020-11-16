package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
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

type DSN struct {
	host      string
	rawurl    string
	key       string
	projectId string
}

func parseDSN(rawurl string) *DSN {
	fmt.Println("> rawlurl", rawurl)

	// TODO support for http vs. https 7: vs 8:
	key := strings.Split(rawurl, "@")[0][7:]

	uri, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	idx := strings.LastIndex(uri.Path, "/")
	if idx == -1 {
		log.Fatal("missing projectId in dsn")
	}
	projectId := uri.Path[idx+1:]

	var host string
	if strings.Contains(rawurl, "ingest.sentry.io") {
		// TODO need to slice the o87286 dynamically
		host = "o87286.ingest.sentry.io"
	}
	if strings.Contains(rawurl, "@localhost:") {
		host = "localhost:9000"
	}
	if host == "" {
		log.Fatal("missing host")
	}
	// if len(key) < 31 || len(key) > 32 {
	// 	log.Fatal("bad key length")
	// }
	if projectId == "" {
		log.Fatal("missing project Id")
	}
	fmt.Printf("> DSN { host: %s, projectId: %s }\n", host, projectId)
	return &DSN{
		host,
		rawurl,
		key,
		projectId,
	}
}

func (d DSN) storeEndpoint() string {
	var fullurl string
	if strings.Contains(d.host, "ingest.sentry.io") {
		// TODO [1:] is for removing leading slash from sentry_key=/a971db611df44a6eaf8993d994db1996, which errors ""bad sentry DSN public key""
		fullurl = fmt.Sprint("https://", d.host, "/api/", d.projectId, "/store/?sentry_key=", d.key[1:], "&sentry_version=7")
		// fullurl = fmt.Sprint("https://", d.host, "/api/", d.projectId, "/store/?sentry_key=", d.key[1:])
	}
	if d.host == "localhost:9000" {
		fullurl = fmt.Sprint("http://", d.host, "/api/", d.projectId, "/store/?sentry_key=", d.key, "&sentry_version=7")
	}
	if fullurl == "" {
		log.Fatal("problem with fullurl")
	}
	return fullurl
}
func (d DSN) envelopeEndpoint() string {
	var fullurl string
	if strings.Contains(d.host, "ingest.sentry.io") {
		fullurl = fmt.Sprint("https://", d.host, "/api/", d.projectId, "/envelope/?sentry_key=", d.key[1:], "&sentry_version=7")
	}
	if d.host == "localhost:9000" {
		fullurl = fmt.Sprint("http://", d.host, "/api/", d.projectId, "/envelope/?sentry_key=", d.key, "&sentry_version=7")
	}
	if fullurl == "" {
		log.Fatal("problem with fullurl")
	}
	return fullurl
}

// TODO put EventJson to its own eventjson.go
type TypeSwitch struct {
	Kind string `json:"type"`
}
type EventJson struct {
	TypeSwitch `json:"type"`
	*Error
	*Transaction
}

func (eventJson *EventJson) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &eventJson.TypeSwitch); err != nil {
		return err
	}
	switch eventJson.Kind {
	case "error":
		eventJson.Error = &Error{}
		return json.Unmarshal(data, eventJson.Error)
	case "transaction":
		eventJson.Transaction = &Transaction{}
		return json.Unmarshal(data, eventJson.Transaction)
	default:
		return fmt.Errorf("unrecognized type value %q", eventJson.Kind)
	}
}

type Event struct {
	Platform string            `json:"platform"`
	Kind     string            `json:"kind"`
	Headers  map[string]string `json:"headers"`
	Body     string            `json:"body"`
}

func (e Event) String() string {
	return fmt.Sprintf("\n Event { Platform: %s, Type: %s }\n", e.Platform, e.Kind) // index somehow?
}
func dsnToStoreEndpoint(projectDSNs map[string]*DSN, projectPlatform string) string {
	// var storeEndpoint string
	if projectPlatform == "javascript" {
		return projectDSNs["javascript"].storeEndpoint()
	} else if projectPlatform == "python" {
		return projectDSNs["python"].storeEndpoint()
	} else {
		log.Fatal("platform type not supported")
	}
	return ""
}
func matchDSN(projectDSNs map[string]*DSN, event Event) string {

	platform := event.Platform

	var storeEndpoint string
	if platform == "javascript" && event.Kind == "error" {
		storeEndpoint = projectDSNs["javascript"].storeEndpoint()
	} else if platform == "python" && event.Kind == "error" {
		storeEndpoint = projectDSNs["python"].storeEndpoint()
	} else if platform == "android" && event.Kind == "error" {
		storeEndpoint = projectDSNs["android"].storeEndpoint()
	} else if platform == "javascript" && event.Kind == "transaction" {
		storeEndpoint = projectDSNs["javascript"].envelopeEndpoint()
	} else if platform == "python" && event.Kind == "transaction" {
		storeEndpoint = projectDSNs["python"].envelopeEndpoint()
	} else if platform == "android" && event.Kind == "transaction" {
		storeEndpoint = projectDSNs["android"].envelopeEndpoint()
	} else {
		log.Fatal("platform type not supported")
	}

	if storeEndpoint == "" {
		log.Fatal("missing store endpoint")
	}
	return storeEndpoint
}

type Timestamp struct {
	time.Time
	rfc3339 bool
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	all = flag.Bool("all", false, "send all events. default is send latest event")
	id = flag.String("id", "", "id of event in sqlite database") // 08/27 non-functional today
	ignore = flag.Bool("i", false, "ignore sending the event to Sentry.io")
	db = flag.String("db", "", "database.json")
	js = flag.String("js", "", "javascript DSN")
	py = flag.String("py", "", "python DSN")
	flag.Parse()

	// sentry +10.0.0 supports performance monitoring, transactions
	projectDSNs = make(map[string]*DSN)
	projectDSNs["javascript"] = parseDSN(os.Getenv("DSN_JAVASCRIPT_SAAS"))
	if *js != "" {
		projectDSNs["javascript"] = parseDSN(*js)
	}
	projectDSNs["python"] = parseDSN(os.Getenv("DSN_PYTHON_SAAS"))
	if *py != "" {
		projectDSNs["python"] = parseDSN(*py)
	}

	if *db == "" {
		database = os.Getenv("JSON")
	} else {
		database = *db
	}
}

func main() {
	// TODO read from CloudStorage
	// jsonFile, err := os.Open(database)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	readJsons(*ignore)
	return

	// CLOUD STORAGE
	bucket := os.Getenv("BUCKET")
	object := database
	fmt.Println("DATASET object", object)
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalln("storage.NewClient:", err)
		return
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		log.Fatalln("NewReader:", err)
		return
	}

	// START
	byteValue, _ := ioutil.ReadAll(rc) // jsonFile
	// defer jsonFile.Close()
	events := make([]Event, 0)
	if err := json.Unmarshal(byteValue, &events); err != nil {
		panic(err)
	}
	requests := []Transport{}
	for _, event := range events {
		fmt.Printf("\n> KIND|PLATFORM %v %v ", event.Kind, event.Platform)

		if event.Kind == "error" {
			bodyError, timestamper, bodyEncoder, storeEndpoint := decodeError(event)
			bodyError = eventId(bodyError)
			bodyError = release(bodyError)
			bodyError = user(bodyError)
			bodyError = timestamper(bodyError, event.Platform)

			requests = append(requests, Transport{
				kind:          event.Kind,
				platform:      event.Platform,
				eventHeaders:  event.Headers,
				storeEndpoint: storeEndpoint,

				bodyError:   bodyError,
				bodyEncoder: bodyEncoder,
			})

		} else if event.Kind == "transaction" {
			envelopeItems, envelopeTimestamper, envelopeEncoder, storeEndpoint := decodeEnvelope(event)
			envelopeItems = eventIds(envelopeItems)
			envelopeItems = envelopeTimestamper(envelopeItems, event.Platform)
			envelopeItems = envelopeReleases(envelopeItems, event.Platform, event.Kind)
			envelopeItems = removeLengthField(envelopeItems)
			envelopeItems = sentAt(envelopeItems)
			envelopeItems = users(envelopeItems)
			getEnvelopeTraceIds(envelopeItems)

			requests = append(requests, Transport{
				kind:          event.Kind,
				platform:      event.Platform,
				eventHeaders:  event.Headers,
				storeEndpoint: storeEndpoint,

				envelopeItems:   envelopeItems,
				envelopeEncoder: envelopeEncoder,
			})
		}

	}

	setEnvelopeTraceIds(requests)
	encodeAndSendEvents(requests, *ignore)

	return
}
