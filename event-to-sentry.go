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
	fmt.Println("> DSN projectId", projectId)

	var host string
	if (strings.Contains(rawurl, "ingest.sentry.io")) {
		host = "ingest.sentry.io" // works with "sentry.io" too
	}
	if (strings.Contains(rawurl, "@localhost:")) {
		host = "localhost:9000"
	}
	fmt.Println("> DSN host", host)
	
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
	return fmt.Sprintf("> event id, type: %v %v", e.id, e._type)
}

func init() {
	defer fmt.Println("> init() complete")
	
	if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
	}

	projects = make(map[string]*DSN)
	
	// projects["javascript"] = parseDSN(os.Getenv("DSN_REACT"))
	// projects["python"] = parseDSN(os.Getenv("DSN_PYTHON"))

	// Must use Hosted Sentry for AM Transactions
	projects["javascript"] = parseDSN(os.Getenv("DSN_REACT_SAAS"))
	projects["python"] = parseDSN(os.Getenv("DSN_PYTHON_SAAS"))

	all = flag.Bool("all", false, "send all events or 1 event from database")
	id = flag.String("id", "", "id of event in sqlite database")
	flag.Parse()
	fmt.Printf("> --all= %v\n", *all)
	fmt.Printf("> --id= %v\n", *id)

	db, _ = sql.Open("sqlite3", "sqlite.db")
}

func javascript(bodyBytes []byte, headers []byte) {
	fmt.Println("> javascript")
	
	bodyInterface := unmarshalJSON(bodyBytes)
	bodyInterface = replaceEventId(bodyInterface)

	// sentry-javascript timestamp is in format "1590946750.683085," https://github.com/getsentry/sentry-javascript/pull/2575
	// undertaker is setting, as it was based on sentry-python which looks like format 2020-05-31T11:10:29.118356Z both formats are supported by Sentry.io
	bodyInterface = addTimestamp(bodyInterface)
	
	bodyBytesPost := marshalJSON(bodyInterface)
	
	SENTRY_URL = projects["javascript"].storeEndpoint()
	fmt.Printf("> storeEndpoint %v", SENTRY_URL)

	request, errNewRequest := http.NewRequest("POST", SENTRY_URL, bytes.NewReader(bodyBytesPost))
	if errNewRequest != nil { log.Fatalln(errNewRequest) }
	
	headerInterface := unmarshalJSON(headers)
	
	// TODO why not just set exactly what came from the database in &event.headers? This way don't have to worry about messing them up...
	// Was it because there were inconsistent headers sent for those python errors earlier? remember - python event.py and Flask error were different
	// "could" make a difference, so maybe hard-code for now??
	// but if can't deduce what kind of tx it is in the proxy, then we wouldn't know here either...
	for _, v := range [4]string{"Accept-Encoding","Content-Length","Content-Type","User-Agent"} {
		request.Header.Set(v, headerInterface[v].(string))
	}
	
	response, requestErr := httpClient.Do(request)
	if requestErr != nil { fmt.Println(requestErr) }

	responseData, responseDataErr := ioutil.ReadAll(response.Body)
	if responseDataErr != nil { log.Fatal(responseDataErr) }

	fmt.Printf("\n> javascript event response\n", responseData)
	// fmt.Printf("\n> javascript event response", string(responseData))
}

func python(bodyBytesCompressed []byte, headers []byte) {
	fmt.Println("> python")
	
	bodyBytes := decodeGzip(bodyBytesCompressed)
	bodyInterface := unmarshalJSON(bodyBytes)
	
	bodyInterface = replaceEventId(bodyInterface)

	// TODO could use addTimestamp? since that's why javascript ^ uses. then re-name it updateTimestamp
	bodyInterface = replaceTimestamp(bodyInterface)
	
	bodyBytesPost := marshalJSON(bodyInterface)
	buf := encodeGzip(bodyBytesPost)
	
	SENTRY_URL = projects["python"].storeEndpoint()
	request, errNewRequest := http.NewRequest("POST", SENTRY_URL, &buf)
	if errNewRequest != nil { log.Fatalln(errNewRequest) }

	headerInterface := unmarshalJSON(headers)

	for _, v := range [5]string{"Accept-Encoding","Content-Length","Content-Encoding","Content-Type","User-Agent"} {
		request.Header.Set(v, headerInterface[v].(string))
	}

	response, requestErr := httpClient.Do(request)
	if requestErr != nil { fmt.Println(requestErr) }

	responseData, responseDataErr := ioutil.ReadAll(response.Body)
	if responseDataErr != nil { log.Fatal(responseDataErr) }

	fmt.Printf("> python event response: %v\n", string(responseData))
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

		if (event._type == "javascript") {
			javascript(event.bodyBytes, event.headers)
		}

		if (event._type == "python") {
			python(event.bodyBytes, event.headers)
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

func replaceTimestamp(bodyInterface map[string]interface{}) map[string]interface{} {
	fmt.Println("before",bodyInterface["timestamp"])
	timestamp := time.Now()
	oldTimestamp := bodyInterface["timestamp"].(string)
	newTimestamp := timestamp.Format("2006-01-02") + "T" + timestamp.Format("15:04:05")
	bodyInterface["timestamp"] = newTimestamp + oldTimestamp[19:]
	fmt.Println("after ",bodyInterface["timestamp"])
	return bodyInterface
}

// SDK's are supposed to set timestamps https://github.com/getsentry/sentry-javascript/issues/2573
// Newer js sdk provides timestamp, so stop calling this function, upon upgrading js sdk. 
func addTimestamp(bodyInterface map[string]interface{}) map[string]interface{} {
	log.Print("no timestamp on object from DB")
	timestamp1 := time.Now()
	newTimestamp1 := timestamp1.Format("2006-01-02") + "T" + timestamp1.Format("15:04:05")
	bodyInterface["timestamp"] = newTimestamp1 + ".118356Z"
	fmt.Println("> after ",bodyInterface["timestamp"])
	return bodyInterface
}