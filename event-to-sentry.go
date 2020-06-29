package main

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	// "github.com/buger/jsonparser"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	// "strconv"
	"strings"
	"time"
)

var httpClient = &http.Client{}

var (
	all         *bool
	id          *string
	ignore      *bool
	db          *sql.DB
	dsn         DSN
	SENTRY_URL  string
	exists      bool
	projectDSNs map[string]*DSN
)

type DSN struct {
	host      string
	rawurl    string
	key       string
	projectId string
}

func parseDSN(rawurl string) *DSN {
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
	if len(key) != 33 {
		log.Fatal("missing key length 33")
	}
	if projectId == "" {
		log.Fatal("missing project Id")
	}
	// fmt.Printf("> DSN { host: %s, projectId: %s }\n", host, projectId)
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

type Event struct {
	id          int
	name, _type string
	headers     []byte
	bodyBytes   []byte
}

func (e Event) String() string {
	return fmt.Sprintf("\n Event { SqliteId: %d, Platform: %s, Type: %s }\n", e.id, e.name, e._type)
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	projectDSNs = make(map[string]*DSN)

	// Must use SAAS for AM Performance Transactions as https://github.com/getsentry/sentry's Release 10.0.0 doesn't include Performance yet
	projectDSNs["javascript"] = parseDSN(os.Getenv("DSN_JAVASCRIPT_SAAS"))
	projectDSNs["python"] = parseDSN(os.Getenv("DSN_PYTHON_SAAS"))
	projectDSNs["node"] = parseDSN(os.Getenv("DSN_EXPRESS_SAAS"))
	projectDSNs["go"] = parseDSN(os.Getenv("DSN_GO_SAAS"))
	projectDSNs["ruby"] = parseDSN(os.Getenv("DSN_RUBY_SAAS"))
	projectDSNs["python_gateway"] = parseDSN(os.Getenv("DSN_PYTHON_GATEWAY"))
	projectDSNs["python_django"] = parseDSN(os.Getenv("DSN_PYTHON_DJANGO"))
	projectDSNs["python_celery"] = parseDSN(os.Getenv("DSN_PYTHON_CELERY"))

	all = flag.Bool("all", false, "send all events or 1 event from database")
	id = flag.String("id", "", "id of event in sqlite database")
	ignore = flag.Bool("i", false, "ignore sending the event to Sentry.io")

	flag.Parse()

	db, _ = sql.Open("sqlite3", os.Getenv("SQLITE"))
}

func jsEncoder(body map[string]interface{}) []byte {
	return marshalJSON(body)
}
func pyEncoder(body map[string]interface{}) []byte {
	bodyBytes := marshalJSON(body)
	buf := encodeGzip(bodyBytes)
	return buf.Bytes()
}

type BodyEncoder func(map[string]interface{}) []byte
type Timestamper func(map[string]interface{}, string) map[string]interface{}

func matchDSN(projectDSNs map[string]*DSN, event Event) string {
	platform := event.name
	headers := unmarshalJSON(event.headers)

	xSentryAuth := headers["X-Sentry-Auth"].(string)
	fmt.Printf("> xSentryAuth %v \n", xSentryAuth)

	for _, projectDSN := range projectDSNs {
		// fmt.Println("projectDSN", keyName, projectDSN.key)
		// TODO remove the leading slash from the key
		if strings.Contains(xSentryAuth, projectDSN.key[1:]) {
			fmt.Println("> match", projectDSN)
			return projectDSN.storeEndpoint()
		}
	}
	// fmt.Println("> event was made by a DSN that was not yours")

	var storeEndpoint string
	if platform == "javascript" {
		storeEndpoint = projectDSNs["javascript"].storeEndpoint()
	} else if platform == "python" {
		storeEndpoint = projectDSNs["python"].storeEndpoint()
	} else {
		log.Fatal("platform type not supported")
	}
	return storeEndpoint
}

func decodeEvent(event Event) (map[string]interface{}, Timestamper, BodyEncoder, []string, string) {
	body := unmarshalJSON(event.bodyBytes)

	JAVASCRIPT := event.name == "javascript"
	PYTHON := event.name == "python"

	ERROR := event._type == "error"
	TRANSACTION := event._type == "transaction"

	// need more discovery on acceptable header combinations by platform/event.type as there seemed to be slight differences in initial testing
	// then could just save the right headers to the database, and not have to deal with this here.
	jsHeaders := []string{"Accept-Encoding", "Content-Length", "Content-Type", "User-Agent"}
	pyHeaders := []string{"Accept-Encoding", "Content-Length", "Content-Encoding", "Content-Type", "User-Agent"}

	storeEndpoint := matchDSN(projectDSNs, event)
	// storeEndpointPython := matchDSN(projectDSNs, event)

	// TODO could check for a run-time DSN mapping file. this way, wouldn't have to bake them into the executable.

	fmt.Printf("> storeEndpoint %T %v \n", storeEndpoint, storeEndpoint)
	// fmt.Printf("> storeEndpointJavascript %T %v ", storeEndpointJavascript, storeEndpointJavascript)

	switch {
	case JAVASCRIPT && TRANSACTION:
		return body, updateTimestamps, jsEncoder, jsHeaders, storeEndpoint //Javascript
	case JAVASCRIPT && ERROR:
		return body, updateTimestamp, jsEncoder, jsHeaders, storeEndpoint
	case PYTHON && TRANSACTION:
		return body, updateTimestamps, pyEncoder, pyHeaders, storeEndpoint
	case PYTHON && ERROR:
		return body, updateTimestamp, pyEncoder, pyHeaders, storeEndpoint
	}

	// TODO need return an error and nil's
	return body, updateTimestamps, jsEncoder, jsHeaders, storeEndpoint
}

func buildRequest(requestBody []byte, headerKeys []string, eventHeaders []byte, storeEndpoint string) *http.Request {
	request, errNewRequest := http.NewRequest("POST", storeEndpoint, bytes.NewReader(requestBody)) // &buf
	if errNewRequest != nil {
		log.Fatalln(errNewRequest)
	}
	headerInterface := unmarshalJSON(eventHeaders)
	for _, v := range headerKeys {
		request.Header.Set(v, headerInterface[v].(string))
	}
	return request
}

func main() {
	defer db.Close()

	query := ""
	if *id == "" {
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

		body, timestamper, bodyEncoder, headerKeys, storeEndpoint := decodeEvent(event)

		body = replaceEventId(body)
		body = timestamper(body, event.name)

		undertake(body)

		requestBody := bodyEncoder(body)
		request := buildRequest(requestBody, headerKeys, event.headers, storeEndpoint)

		if !*ignore {
			response, requestErr := httpClient.Do(request)
			if requestErr != nil {
				fmt.Println(requestErr)
			}

			responseData, responseDataErr := ioutil.ReadAll(response.Body)
			if responseDataErr != nil {
				log.Fatal(responseDataErr)
			}

			fmt.Printf("> %s event response %s\n", event._type, string(responseData))
		} else {
			fmt.Printf("> %s event IGNORED", event._type)
		}

		if !*all {
			rows.Close()
		}

		time.Sleep(300 * time.Millisecond)
	}
	rows.Close()
}

func replaceEventId(bodyInterface map[string]interface{}) map[string]interface{} {
	if _, ok := bodyInterface["event_id"]; !ok {
		log.Print("no event_id on object from DB")
	}
	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
	bodyInterface["event_id"] = uuid4
	fmt.Println("> event_id updated", bodyInterface["event_id"])
	return bodyInterface
}

// Python Error Events do not have 'tags' attribute, if no custom tags were set...? "Sometimes there's no tags attribute yet (typically if no custom tags were set, at least for ERr EVents". Transactions come with a few tags by default, by the sdk.
func undertake(bodyInterface map[string]interface{}) {
	if bodyInterface["tags"] == nil {
		bodyInterface["tags"] = make(map[string]interface{})
	}
	tags := bodyInterface["tags"].(map[string]interface{})
	tags["undertaker"] = "crontab"

	// Optional - overwrite the platform (make sure matches the DSN's project type)
	// bodyInterface["platform"] = "ruby"
	// Optional - overwrite what the transaction's title will display as in Discover
	// bodyInterface["transaction"] = "eprescription/:id"
}

////////////////////////////  UTILS  /////////////////////////////////////////
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
	if errBodyBytes != nil {
		fmt.Println(errBodyBytes)
	}
	return bodyBytes
}
