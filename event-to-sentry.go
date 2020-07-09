package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	// "github.com/buger/jsonparser"
	"io/ioutil"
	"log"
	"math/rand"
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
	database	*sql.DB
	db			*string
	js			*string
	py			*string
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
	// key := strings.Split(rawurl, "@")[0][7:]
	key := strings.Split(rawurl, "@")[0][8:]

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
	if len(key) != 32 {
		log.Fatal("missing key length 32")
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
	platform, _type string
	headers     []byte
	bodyBytes   []byte
}

func (e Event) String() string {
	return fmt.Sprintf("\n Event { SqliteId: %d, Platform: %s, Type: %s }\n", e.id, e.platform, e._type)
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
	platform := event.platform
	headers := unmarshalJSON(event.headers)

	// only python events have X-Sentry-Auth
	if headers["X-Sentry-Auth"] != nil {
		xSentryAuth := headers["X-Sentry-Auth"].(string)
		for _, projectDSN := range projectDSNs {
			if strings.Contains(xSentryAuth, projectDSN.key) {
				fmt.Println("> match", projectDSN)
				return projectDSN.storeEndpoint()
			}
		}
	}
	
	// event was made by a DSN that was not yours, so we can't match it, use default javascript/python DSN in .env
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

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	all = flag.Bool("all", false, "send all events or 1 event from database")
	id = flag.String("id", "", "id of event in sqlite database")
	ignore = flag.Bool("i", false, "ignore sending the event to Sentry.io")
	db = flag.String("db", "", "path-to-database.db")
	js = flag.String("js", "", "javascript DSN")
	py = flag.String("py", "", "python DSN")
	flag.Parse()

	// Use SAAS DSN's for Tx's as getsentry/sentry 10.0.0 doesn't include Tx's yet
	projectDSNs = make(map[string]*DSN)
	projectDSNs["javascript"] = parseDSN(os.Getenv("DSN_JAVASCRIPT_SAAS"))
	if (*js != "") {
		projectDSNs["javascript"] = parseDSN(*js)
	}
	projectDSNs["python"] = parseDSN(os.Getenv("DSN_PYTHON_SAAS"))
	if (*py != "") {
		projectDSNs["python"] = parseDSN(*py)
	}
	projectDSNs["node"] = parseDSN(os.Getenv("DSN_EXPRESS_SAAS"))
	projectDSNs["go"] = parseDSN(os.Getenv("DSN_GO_SAAS"))
	projectDSNs["ruby"] = parseDSN(os.Getenv("DSN_RUBY_SAAS"))
	projectDSNs["python_gateway"] = parseDSN(os.Getenv("DSN_PYTHON_GATEWAY"))
	projectDSNs["python_django"] = parseDSN(os.Getenv("DSN_PYTHON_DJANGO"))
	projectDSNs["python_celery"] = parseDSN(os.Getenv("DSN_PYTHON_CELERY"))

	fmt.Println("> db flag", *db)
	if *db == "" {
		database, _ = sql.Open("sqlite3", os.Getenv("SQLITE"))
	} else {
		database, _ = sql.Open("sqlite3", *db)
	}
}

func main() {
	defer database.Close()

	query := ""
	if *id == "" {
		query = "SELECT * FROM events ORDER BY id DESC"
	} else {
		query = strings.ReplaceAll("SELECT * FROM events WHERE id=?", "?", *id)
	}

	rows, err := database.Query(query)

	if err != nil {
		fmt.Println("Failed to load rows", err)
	}
	for rows.Next() {
		var event Event
		rows.Scan(&event.id, &event.platform, &event._type, &event.bodyBytes, &event.headers)
		fmt.Println(event)

		body, timestamper, bodyEncoder, headerKeys, storeEndpoint := decodeEvent(event)

		body = eventId(body)
		body = release(body)
		body = user(body)
		body = timestamper(body, event.platform)

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

			fmt.Printf("\n> event type: %s, response: %s\n", event._type, string(responseData))
		} else {
			fmt.Printf("\n> %s event IGNORED", event._type)
		}

		if !*all {
			rows.Close()
		}

		time.Sleep(1000 * time.Millisecond)
	}
	rows.Close()
}

func decodeEvent(event Event) (map[string]interface{}, Timestamper, BodyEncoder, []string, string) {
	body := unmarshalJSON(event.bodyBytes)

	JAVASCRIPT := event.platform == "javascript"
	PYTHON := event.platform == "python"

	ERROR := event._type == "error"
	TRANSACTION := event._type == "transaction"

	// need more discovery on acceptable header combinations by platform/event.type as there seemed to be slight differences in initial testing
	// then could just save the right headers to the database, and not have to deal with this here.
	jsHeaders := []string{"Accept-Encoding", "Content-Length", "Content-Type", "User-Agent"}
	pyHeaders := []string{"Accept-Encoding", "Content-Length", "Content-Encoding", "Content-Type", "User-Agent"}

	storeEndpoint := matchDSN(projectDSNs, event)

	fmt.Printf("> storeEndpoint %v \n", storeEndpoint)

	switch {
	case JAVASCRIPT && TRANSACTION:
		return body, updateTimestamps, jsEncoder, jsHeaders, storeEndpoint
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

// same eventId cannot be accepted twice by Sentry
func eventId(body map[string]interface{}) map[string]interface{} {
	if _, ok := body["event_id"]; !ok {
		log.Print("no event_id on object from DB")
	}
	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
	body["event_id"] = uuid4
	fmt.Println("> event_id updated", body["event_id"])
	return body
}

// CalVer-lite
func release(body map[string]interface{}) map[string]interface{} {
	date := time.Now()
	day := date.Day()
	var week int
	switch {
	case day <= 7:
		week = 1
	case day >= 8 && day <= 14:
		week = 2
	case day >= 15 && day <= 21:
		week = 3
	case day >= 22:
		week = 4
	}
	release := fmt.Sprint(int(date.Month()), ".", week)
	body["release"] = release
	fmt.Println("> release updated", body["release"])
	return body
}

// if it's a back-end event, this randomly generated user will not match the user from the corresponding front end (trace) event
// so it's better to never miss setting the user from the SDK
func user(body map[string]interface{}) map[string]interface{} {
	if body["user"] == nil {
		body["user"] = make(map[string]interface{})
		user := body["user"].(map[string]interface{})
		rand.Seed(time.Now().UnixNano())
		alpha := strings.Split("abcdefghijklmnopqrstuvwxyz", "")[rand.Intn(9)]
		var alphanumeric string
		for i := 0; i < 3; i++ {
			alphanumeric += strings.Split("abcdefghijklmnopqrstuvwxyz0123456789", "")[rand.Intn(35)]
		}
		user["email"] = fmt.Sprint(alpha, alphanumeric, "@yahoo.com")
	}
	// fmt.Println("> user", body["user"])
	return body
}

func undertake(body map[string]interface{}) {
	if body["tags"] == nil {
		body["tags"] = make(map[string]interface{})
	}
	tags := body["tags"].(map[string]interface{})
	tags["undertaker"] = "crontab"
}
