package main

import (
	"bytes"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"github.com/joho/godotenv"
	"strings"
	"time"
	"encoding/json"
)

var httpClient = &http.Client{}

var (
	all         *bool
	id          *string
	ignore      *bool
	database    string
	db			*string
	js			*string
	py			*string
	dsn         DSN
	SENTRY_URL  string
	exists      bool
	projectDSNs map[string]*DSN
	// traceIdMap0 map[string][]*Item
	traceIdMap map[string][]interface{}
	
	// traceIdMap2 map[string][]string
	// traceIdMap := map[string][]*interface{}
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
	Platform    string `json:"platform"`
	Kind        string `json:"kind"`
	Headers     map[string]string `json:"headers"`
	Body        string `json:"body"`
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
	Sent_at string `json:"sent_at,omitempty"`

	Length int `json:"length,omitempty"`
	Type string `json:"type,omitempty"`
	Content_type string `json:"content_type,omitempty"`

	Start_timestamp string `json:"start_timestamp,omitempty"`
	Transaction string `json:"transaction,omitempty"`
	Server_name string `json:"server_name,omitempty"`
	Tags map[string]interface{} `json:"tags,omitempty"`
	Contexts map[string]interface{} `json:"contexts,omitempty"`
	

	Extra map[string]interface{} `json:"extra,omitempty"`
	Request map[string]interface{} `json:"request,omitempty"`
	Environment string `json:"environment,omitempty"`
	Platform string `json:"platform,omitempty"`
	// Todo spans []
	Sdk map[string]interface{} `json:"sdk,omitempty"`
	User map[string]interface{} `json:"user,omitempty"`
}

// TODO need an ItemFinal that has unified timestamp?

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	traceIdMap = make(map[string][]interface{})

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
	if (*js != "") {
		projectDSNs["javascript"] = parseDSN(*js)
	}
	projectDSNs["python"] = parseDSN(os.Getenv("DSN_PYTHON_SAAS"))
	if (*py != "") {
		projectDSNs["python"] = parseDSN(*py)
	}

	if *db == "" {
		database = os.Getenv("JSON")
	} else {
		database = *db
	}
}

func main() {
	jsonFile, err := os.Open(database)
	if err != nil {
		log.Fatal(err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	defer jsonFile.Close()
	events := make([]Event, 0)
	if err := json.Unmarshal(byteValue, &events); err != nil {
		panic(err)
	}

	for _, event := range events {
		fmt.Printf("\n> KIND|PLATFORM %v %v ", event.Kind, event.Platform)
		var bodyError map[string]interface{}
		var envelopeItems []interface{}

		var timestamper Timestamper 
		var envelopeTimestamper EnvelopeTimestamper
		var bodyEncoder BodyEncoder
		var envelopeEncoder EnvelopeEncoder
		var storeEndpoint string
		var requestBody []byte

		if (event.Kind == "error") {			
			bodyError, timestamper, bodyEncoder, storeEndpoint = decodeError(event)
			bodyError = eventId(bodyError)
			bodyError = release(bodyError)
			bodyError = user(bodyError)
			bodyError = timestamper(bodyError, event.Platform)
			undertake(bodyError)
			requestBody = bodyEncoder(bodyError)
		} else if (event.Kind == "transaction") {
			envelopeItems, envelopeTimestamper, envelopeEncoder, storeEndpoint = decodeEnvelope(event)
			envelopeItems = eventIds(envelopeItems)
			envelopeItems = envelopeTimestamper(envelopeItems, event.Platform)
			envelopeItems = envelopeReleases(envelopeItems, event.Platform, event.Kind)
			envelopeItems = getEnvelopeTraceIds(envelopeItems)

			// update user if missing
			envelopeItems = removeLengthField(envelopeItems)
			requestBody = envelopeEncoder(envelopeItems)
			// undertaker()			
		}

		// 1. Array of Envelopes
			// encode later, after setting new TraceId
			// build Request later

		request := buildRequest(requestBody, event.Headers, storeEndpoint)
		if !*ignore {
			response, requestErr := httpClient.Do(request)
			if requestErr != nil {
				log.Fatal(requestErr)
			}
			responseData, responseDataErr := ioutil.ReadAll(response.Body)
			if responseDataErr != nil {
				log.Fatal(responseDataErr)
			}
			fmt.Printf("> KIND|RESPONSE: %s %s\n", event.Kind, string(responseData))
		} else {
			fmt.Printf("> %s event IGNORED \n", event.Kind)
		}

		// break early, or auto-select 1 before the for loop
		// if !*all {
		// 	return
		// }

		time.Sleep(1000 * time.Millisecond)
	}
	// 2. Update All Envelopes with proper Trace Id - setEnvelopeTraceIds
	// 3. Send all to Sentry.io
		// build all requests, or have that done ahead of time.
	return
}

func buildRequest(requestBody []byte, eventHeaders map[string]string, storeEndpoint string) *http.Request {
	fmt.Printf("> storeEndpoint %v \n", storeEndpoint)
	if requestBody == nil {
		log.Fatalln("buildRequest missing requestBody")
	}
	if eventHeaders == nil {
		log.Fatalln("buildRequest missing eventHeaders")
	}
	if storeEndpoint == "" {
		log.Fatalln("buildRequest missing storeEndpoint")
	}

	request, errNewRequest := http.NewRequest("POST", storeEndpoint, bytes.NewReader(requestBody)) // &buf
	if errNewRequest != nil {
		log.Fatalln(errNewRequest)
	}

	for key, value := range eventHeaders {
		if (key != "X-Sentry-Auth") {
			request.Header.Set(key, value)
		}
	}
	return request
}

