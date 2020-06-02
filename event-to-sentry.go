package main

import (
	_ "github.com/mattn/go-sqlite3"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	// "github.com/buger/jsonparser"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var httpClient = &http.Client{}

var (
	all *bool
	id *string
	db *sql.DB
	dsn DSN
	SENTRY_URL string 
	exists bool
	projects map[string]*DSN
)

type DSN struct { 
	host string
	rawurl string
	key string
	projectId string
}

func parseDSN(rawurl string) (*DSN) {
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
	// fmt.Println("> DSN projectId", projectId)
	
	var host string
	if (strings.Contains(rawurl, "ingest.sentry.io")) {
		host = "ingest.sentry.io" // works with "sentry.io" too?
	}
	if (strings.Contains(rawurl, "@localhost:")) {
		host = "localhost:9000"
	}

	fmt.Printf("> DSN { host: %s, projectId: %s }\n", host, projectId)
	
	return &DSN{
		host,
		rawurl,
		key,
		projectId,
	}
}

// TODO could make a DSN field called 'storeEndpoint' and use this function there to assign the value, during parseDSN
func (d DSN) storeEndpoint() string {
	var fullurl string
	if (d.host == "ingest.sentry.io") {
		// still works if you pass in the "o87286"
		// fullurl = fmt.Sprint("https://o87286.",d.host,"/api/",d.projectId,"/store/?sentry_key=",d.key,"&sentry_version=7")
		fullurl = fmt.Sprint("https://",d.host,"/api/",d.projectId,"/store/?sentry_key=",d.key,"&sentry_version=7")
	}
	if (d.host == "localhost:9000") {
		fullurl = fmt.Sprint("http://",d.host,"/api/",d.projectId,"/store/?sentry_key=",d.key,"&sentry_version=7")
	}
	return fullurl
}

type Event struct {
	id int
	name, _type string
	headers []byte
	bodyBytes []byte
}
func (e Event) String() string {
	return fmt.Sprintf("\n Event { SqliteId: %d, Platform: %s, Type: %s }\n", e.id, e.name, e._type)
}

func init() {
	if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
	}

	projects = make(map[string]*DSN)
	
	// Must use Hosted Sentry for AM Transactions
	// projects["javascript"] = parseDSN(os.Getenv("DSN_REACT"))
	// projects["python"] = parseDSN(os.Getenv("DSN_PYTHON"))
	projects["javascript"] = parseDSN(os.Getenv("DSN_REACT_SAAS"))
	projects["python"] = parseDSN(os.Getenv("DSN_PYTHONTEST_SAAS"))

	all = flag.Bool("all", false, "send all events or 1 event from database")
	id = flag.String("id", "", "id of event in sqlite database")
	flag.Parse()
	// fmt.Printf("> --all= %v\n", *all)
	// fmt.Printf("> --id= %v\n", *id)

	db, _ = sql.Open("sqlite3", "am-transactions-sqlite.db")
}

func javascript(event Event) {
	fmt.Sprintf("> JAVASCRIPT %v %v", event.name, event._type)
	
	bodyInterface := unmarshalJSON(event.bodyBytes)
	bodyInterface = replaceEventId(bodyInterface)

	if (event._type == "error") {
		bodyInterface = updateTimestamp(bodyInterface, "javascript")
	}
	if (event._type == "transaction") {
		bodyInterface = updateTimestamps(bodyInterface, "javascript")
	}

	bodyBytesPost := marshalJSON(bodyInterface)
	
	SENTRY_URL = projects["javascript"].storeEndpoint()
	fmt.Printf("> storeEndpoint %v", SENTRY_URL)

	request, errNewRequest := http.NewRequest("POST", SENTRY_URL, bytes.NewReader(bodyBytesPost))
	if errNewRequest != nil { log.Fatalln(errNewRequest) }
	
	headerInterface := unmarshalJSON(event.headers)
	
	for _, v := range [4]string{"Accept-Encoding","Content-Length","Content-Type","User-Agent"} {
		request.Header.Set(v, headerInterface[v].(string))
	}
	
	response, requestErr := httpClient.Do(request)
	if requestErr != nil { fmt.Println(requestErr) }

	responseData, responseDataErr := ioutil.ReadAll(response.Body)
	if responseDataErr != nil { log.Fatal(responseDataErr) }

	// TODO this prints nicely if response is coming from Self-Hosted. Not the case when sending to Hosted sentry
	fmt.Printf("\n> javascript event response", string(responseData))
}

func python(event Event) {
	fmt.Sprintf("> PYTHON %v %v", event.name, event._type)
	// bodyBytes := decodeGzip(bodyBytesCompressed)
	bodyInterface := unmarshalJSON(event.bodyBytes)
	bodyInterface = replaceEventId(bodyInterface)

	// updateTimestamp 2020-05-31T11:10:29.118356Z
	if (event._type == "error") {
		bodyInterface = updateTimestamp(bodyInterface, "python")
	}
	if (event._type == "transaction") {
		bodyInterface = updateTimestamps(bodyInterface, "python")
	}
	
	// fmt.Printf("> timestamp %v\n", bodyInterface["timestamp"])
	
	bodyBytesPost := marshalJSON(bodyInterface)
	buf := encodeGzip(bodyBytesPost)
	
	SENTRY_URL = projects["python"].storeEndpoint()
	fmt.Printf("> storeEndpoint %v", SENTRY_URL)

	request, errNewRequest := http.NewRequest("POST", SENTRY_URL, &buf)
	if errNewRequest != nil { log.Fatalln(errNewRequest) }

	headerInterface := unmarshalJSON(event.headers)

	// X-Sentry-Auth
	for _, v := range [6]string{"Accept-Encoding","Content-Length","Content-Encoding","Content-Type","User-Agent", "X-Sentry-Auth"} {
		request.Header.Set(v, headerInterface[v].(string))
	}

	response, requestErr := httpClient.Do(request)
	if requestErr != nil { fmt.Println(requestErr) }

	responseData, responseDataErr := ioutil.ReadAll(response.Body)
	if responseDataErr != nil { log.Fatal(responseDataErr) }

	fmt.Printf("\n> python event response: %v\n", string(responseData))
}

func main() {
	defer db.Close()
	
	query := ""
	if (*id == "") {
		query = "SELECT * FROM events ORDER BY id DESC"
	} else {
		query = strings.ReplaceAll("SELECT * FROM events WHERE id=?", "?", *id)
	}

	rows, err := db.Query(query)
	
	if err != nil {
		fmt.Println("Failed to load rows", err)
	}
	for rows.Next() {
		var event Event
		rows.Scan(&event.id, &event.name, &event._type, &event.bodyBytes, &event.headers)
		fmt.Println(event)

		if (event.name == "javascript") {
			javascript(event)
		}

		if (event.name == "python") {
			python(event)
		}

		if !*all {
			rows.Close()
		}
	}
	rows.Close()
}

func decodeGzip(bodyBytesInput []byte) (bodyBytesOutput []byte) {
	bodyReader, err := gzip.NewReader(bytes.NewReader(bodyBytesInput))
	if err != nil {
		fmt.Println(err)
	}
	bodyBytesOutput, err = ioutil.ReadAll(bodyReader)
	if err != nil {
		fmt.Println(err)
	}
	return
}

func encodeGzip(b []byte) bytes.Buffer {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(b)
	w.Close()
	// return buf.Bytes()
	return buf
}

func unmarshalJSON(bytes []byte) map[string]interface{} {
	var _interface map[string]interface{}
	if err := json.Unmarshal(bytes, &_interface); err != nil {
		panic(err)
	}
	return _interface
}

func marshalJSON(bodyInterface map[string]interface{}) []byte {
	bodyBytes, errBodyBytes := json.Marshal(bodyInterface) 
	if errBodyBytes != nil { fmt.Println(errBodyBytes)}
	return bodyBytes
}

func replaceEventId(bodyInterface map[string]interface{}) map[string]interface{} {
	if _, ok := bodyInterface["event_id"]; !ok { 
		log.Print("no event_id on object from DB")
	}

	fmt.Println("> before",bodyInterface["event_id"])
	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "") 
	bodyInterface["event_id"] = uuid4
	fmt.Println("> after ",bodyInterface["event_id"])
	return bodyInterface
}

// js timestamps https://github.com/getsentry/sentry-javascript/pull/2575
func updateTimestamp(bodyInterface map[string]interface{}, platform string) map[string]interface{} {
	fmt.Println(" timestamp before", bodyInterface["timestamp"]) // nil for js errors, despite being on latest sdk as of 05/30/2020
	
	// "1590946750" unix
	if (platform == "javascript") {
		bodyInterface["timestamp"] = time.Now().Unix() 
	}

	// "2020-05-31T23:55:11.807534Z"
	if (platform == "python") {
		// is PST, or wherever you're running this from
		timestamp := time.Now()
		// is GMT, so not same as timezone you're running this from
		oldTimestamp := bodyInterface["timestamp"].(string)
		newTimestamp := timestamp.Format("2006-01-02") + "T" + timestamp.Format("15:04:05")
		bodyInterface["timestamp"] = newTimestamp + oldTimestamp[19:]

		// TODO these need to match. 'timestamp before' is GTC, appearing as far ahead of PST.
		// timestamp before 2020-06-02T00:09:51.365214Z
		// timestamp after 2020-06-01T17:12:26.365214Z
	   
		// doesn't work, won't appear in Sentry.io
		// bodyInterface["timestamp"] = time.Now().Unix()
	}

	fmt.Println("  timestamp after", bodyInterface["timestamp"]) // nil for js errors, despite being on latest sdk as of 05/30/2020
	return bodyInterface
}

func updateTimestamps(bodyInterface map[string]interface{}, platform string) map[string]interface{} {
	// 'start_timestamp' is only present in transactions
	fmt.Println("       timestamp before",bodyInterface["start_timestamp"])
	fmt.Println(" start_timestamp before",bodyInterface["start_timestamp"])
	// event.context.trace.span_id
	// event.startTimestamp
	// event.endTimestamp

	// for span in event.entrices[0].data: 
		// start_timestamp
		// timestamp

	fmt.Println("       timestamp after",bodyInterface["start_timestamp"])
	fmt.Println(" start_timestamp after",bodyInterface["start_timestamp"])
	
	return bodyInterface
}

// SDK's are supposed to set timestamps https://github.com/getsentry/sentry-javascript/issues/2573
// Newer js sdk provides timestamp, so stop calling this function, upon upgrading js sdk. 
// func addTimestamp(bodyInterface map[string]interface{}) map[string]interface{} {
// 	log.Print("no timestamp on object from DB")
	
// 	timestamp1 := time.Now()
// 	newTimestamp1 := timestamp1.Format("2006-01-02") + "T" + timestamp1.Format("15:04:05")
// 	bodyInterface["timestamp"] = newTimestamp1 + ".118356Z"

// 	// bodyInterface["timestamp"] = "1590957221.4570072"
// 	fmt.Println("> after ",bodyInterface["timestamp"])
// 	return bodyInterface
// }