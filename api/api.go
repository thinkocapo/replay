package undertaker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/getsentry/sentry-go"
	// _ "github.com/mattn/go-sqlite3"
	// "github.com/getsentry/sentry-go"
	// sentryhttp "github.com/getsentry/sentry-go/http"
)

var httpClient = &http.Client{}

var (
	all         *bool
	id          *string
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
		host = "ingest.sentry.io"
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
	if d.host == "ingest.sentry.io" {
		fullurl = fmt.Sprint("https://", d.host, "/api/", d.projectId, "/store/?sentry_key=", d.key, "&sentry_version=7")
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
	if d.host == "ingest.sentry.io" {
		fullurl = fmt.Sprint("https://", d.host, "/api/", d.projectId, "/envelope/?sentry_key=", d.key, "&sentry_version=7")
	}
	if d.host == "localhost:9000" {
		fullurl = fmt.Sprint("http://", d.host, "/api/", d.projectId, "/envelope/?sentry_key=", d.key, "&sentry_version=7")
	}
	if fullurl == "" {
		log.Fatal("problem with fullurl")
	}
	return fullurl
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

// type Envelope struct {
// 	items []interface{}
// }

// type Item map[string]interface{}

// type Timestamp time.Time
type Timestamp struct {
	time.Time
	rfc3339 bool
}

type Item struct {
	Timestamp Timestamp `json:"timestamp,omitempty"`
	// Timestamp time.Time `json:"timestamp,omitempty"`

	Event_id string `json:"event_id,omitempty"`
	Sent_at  string `json:"sent_at,omitempty"`

	Length       int    `json:"length,omitempty"`
	Type         string `json:"type,omitempty"`
	Content_type string `json:"content_type,omitempty"`

	Start_timestamp string                 `json:"start_timestamp,omitempty"`
	Transaction     string                 `json:"transaction,omitempty"`
	Server_name     string                 `json:"server_name,omitempty"`
	Tags            map[string]interface{} `json:"tags,omitempty"`
	Contexts        map[string]interface{} `json:"contexts,omitempty"`

	Extra       map[string]interface{} `json:"extra,omitempty"`
	Request     map[string]interface{} `json:"request,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Platform    string                 `json:"platform,omitempty"`
	// Todo spans []
	Sdk  map[string]interface{} `json:"sdk,omitempty"`
	User map[string]interface{} `json:"user,omitempty"`
}

// TODO need an ItemFinal that has unified timestamp?

func init() {
	fmt.Print("Init...123")
	// err := sentry.Init(sentry.ClientOptions{
	// 	Dsn: "https://879a3ddfdd5241b0b4f6fcf9011896ad@o87286.ingest.sentry.io/5426957",
	// 	Debug: false,
	// })
	// if err != nil {
	// 	log.Fatalf("sentry.Init: %s", err)
	// }
	// defer sentry.Flush(2 * time.Second)
}

type DsnPair struct {
	Dsn1 string `json:"dsn1"`
	Dsn2 string `json:"dsn2"`
}

func ReplayJson(w http.ResponseWriter, r *http.Request) {
	// sentryHandler := sentryhttp.New(sentryhttp.Options{})
	// sentry.CaptureMessage("SDK WORKS")

	var d DsnPair

	err0 := json.NewDecoder(r.Body).Decode(&d)
	if err0 != nil {
		sentry.CaptureException(err0)
		http.Error(w, err0.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("dsn", d)
	dsn1 := d.Dsn1
	dsn2 := d.Dsn2

	dsn := r.Header.Get("dsn") // py default for just 1 python error
	// dsn1 := r.Header.Get("dsn1") // js
	// dsn2 := r.Header.Get("dsn2") // py
	fmt.Println("dsn1 is", dsn1)
	fmt.Println("dsn2 is", dsn2)

	if dsn1 == "" && dsn2 == "" {
		fmt.Fprint(w, "missing a DSN key")
		return
	}

	projectDSNs = make(map[string]*DSN)
	projectDSNs["javascript"] = parseDSN(dsn1)
	if dsn != "" {
		projectDSNs["python"] = parseDSN(dsn)
	} else {
		projectDSNs["python"] = parseDSN(dsn2)
	}

	// CLOUD STORAGE
	bucket := os.Getenv("BUCKET")
	object := "npm-sentry-tracing.json"
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
	byteValue, _ := ioutil.ReadAll(rc)
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
				bodyError:     bodyError,
				bodyEncoder:   bodyEncoder,
			})

		} else if event.Kind == "transaction" {

			envelopeItems, envelopeTimestamper, envelopeEncoder, storeEndpoint := decodeEnvelope(event)
			envelopeItems = eventIds(envelopeItems)
			envelopeItems = envelopeTimestamper(envelopeItems, event.Platform)
			envelopeItems = envelopeReleases(envelopeItems, event.Platform, event.Kind)
			envelopeItems = removeLengthField(envelopeItems)

			getEnvelopeTraceIds(envelopeItems)

			requests = append(requests, Transport{
				kind:            event.Kind,
				platform:        event.Platform,
				eventHeaders:    event.Headers,
				storeEndpoint:   storeEndpoint,
				envelopeItems:   envelopeItems,
				envelopeEncoder: envelopeEncoder,
			})
		}

	}

	setEnvelopeTraceIds(requests)
	encodeAndSendEvents(requests)

	message := "DONE" + strconv.Itoa(len(requests)) + "requests"
	fmt.Fprint(w, message)
	// fmt.Fprint(w, "DONE %s %s", "requests", "sdf")
}
