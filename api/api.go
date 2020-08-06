package undertaker

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"log"
	"github.com/google/uuid"
	"net/url"
	"strings"
	"time"
	"cloud.google.com/go/storage"
	"math/rand"
	"os"
)

var httpClient = &http.Client{}

var (
	projectDSNs map[string]*DSN
)

type DSN struct {
	host      string
	rawurl    string
	key       string
	projectId string
}

func parseDSN(rawurl string) *DSN {
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
	Platform    string `json:"platform"`
	Kind        string `json:"kind"`
	Headers     map[string]string `json:"headers"`
	Body map[string]interface{} `json:"body"`
}

func (e Event) String() string {
	return fmt.Sprintf("\n Event { Platform: %s, Type: %s }\n", e.Platform, e.Kind) // index somehow?
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
	platform := event.Platform
	headers := event.Headers

	if headers["X-Sentry-Auth"] != "" {
		xSentryAuth := headers["X-Sentry-Auth"]
		for _, projectDSN := range projectDSNs {
			if strings.Contains(xSentryAuth, projectDSN.key) {
				fmt.Println("> match", projectDSN)
				return projectDSN.storeEndpoint()
			}
		}
	}
	
	var storeEndpoint string
	if platform == "javascript" {
		storeEndpoint = projectDSNs["javascript"].storeEndpoint()
	} else if platform == "python" {
		storeEndpoint = projectDSNs["python"].storeEndpoint()
	} else if platform == "android" {
		storeEndpoint = projectDSNs["android"].storeEndpoint()
	} else {
		log.Fatal("platform type not supported")
	}
	return storeEndpoint
}

func init() {
	fmt.Print("Init...")
	// projectDSNs["python_gateway"] = parseDSN("")
	// projectDSNs["python_django"] = parseDSN("")
	// projectDSNs["python_celery"] = parseDSN("")
}

func Replay(w http.ResponseWriter, r *http.Request) {
	bucket := os.Getenv("BUCKET")

	dsn := r.Header.Get("dsn") // py default for just 1 python error
	dsn1 := r.Header.Get("dsn1") // js
	dsn2 := r.Header.Get("dsn2") // py
	fmt.Println("dsn", dsn)
	fmt.Println("dsn1", dsn1)
	fmt.Println("dsn2", dsn2)

	if (dsn == "" && dsn1 == "" && dsn2 == "") {
		fmt.Fprint(w, "no DSN key provided")
		return
	}
	fmt.Println("I SHOULD NOT LOG IF NO DSN WAS PROVIDED")

	projectDSNs = make(map[string]*DSN)
	projectDSNs["javascript"] = parseDSN(dsn1)
	if dsn != "" {
		projectDSNs["python"] = parseDSN(dsn)
	} else {
		projectDSNs["python"] = parseDSN(dsn2)
	}

	// Dataset
	var object string
	DATASET := r.Header.Get("data")
	if DATASET == "" {
		object = "eventsa.json"
	} else {
		object = DATASET
	}
	fmt.Println("DATASET object", object)

	// TODO move context.Background() to init function...?
	// if *id == ""
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Fprint(w, "storage.NewClient: %v", err)
		return
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		fmt.Fprint(w, "NewReader: %v", err)
		return
	}
	data, err := ioutil.ReadAll(rc)
	if err != nil {
		fmt.Fprint(w, "ioutil.ReadAll: %v", err)
		return
	}
	events := make([]Event, 0)
	if err := json.Unmarshal(data, &events); err != nil {
		fmt.Fprint(w, "couldn't unmarshal data: %v", err)
		return
	}

	fmt.Println("> EVENTS length", len(events))
	for idx, event := range events {
		fmt.Printf("> EVENT # %v \n", idx)

		var body map[string]interface{}
		var timestamper Timestamper 
		var bodyEncoder BodyEncoder
		var headerKeys []string
		var storeEndpoint string
		var requestBody []byte

		body, timestamper, bodyEncoder, headerKeys, storeEndpoint = decodeEvent(event)
		body = eventId(body)
		body = release(body)
		body = user(body)
		body = timestamper(body, event.Platform)
		
		undertake(body)
		requestBody = bodyEncoder(body)
		request := buildRequest(requestBody, headerKeys, event.Headers, storeEndpoint)

		// 'ignore' is for skipping the final call to Sentry
		ignore := ""	
		if ignore == "" {
			response, requestErr := httpClient.Do(request)
			if requestErr != nil {
				fmt.Fprint(w, "httpClient.Do(request) failed: %v", requestErr)
				return
			}
			responseData, responseDataErr := ioutil.ReadAll(response.Body)
			if responseDataErr != nil {
				fmt.Fprint(w, "error in responseData: %v", responseDataErr)
				return
			}

			fmt.Printf("\n> EVENT KIND: %s | RESPONSE: %s\n", event.Kind, string(responseData))
			fmt.Fprint(w, "\n> EVENT made: ", event.Kind, string(responseData))
		} else {
			fmt.Printf("\n> %s event IGNORED", event.Kind)
		}

		time.Sleep(500 * time.Millisecond)
	}
	fmt.Fprint(w, "\n>FINISHED all - go check Sentry")
}

func decodeEvent(event Event) (map[string]interface{}, Timestamper, BodyEncoder, []string, string) {
	body := event.Body

	JAVASCRIPT := event.Platform == "javascript"
	PYTHON := event.Platform == "python"
	ANDROID := event.Platform == "android"
	
	ERROR := event.Kind == "error"
	TRANSACTION := event.Kind == "transaction"
	
	jsHeaders := []string{"Accept-Encoding", "Content-Length", "Content-Type", "User-Agent"}
	pyHeaders := []string{"Accept-Encoding", "Content-Length", "Content-Encoding", "Content-Type", "User-Agent"}
	androidHeaders := []string{"Content-Length","User-Agent","Connection","Content-Encoding","X-Forwarded-Proto","Host","Accept","X-Forwarded-For"} // X-Sentry-Auth omitted
	
	storeEndpoint := matchDSN(projectDSNs, event)
	fmt.Printf("> storeEndpoint %v \n", storeEndpoint)

	switch {
	case ANDROID && TRANSACTION:
		return body, updateTimestamp, pyEncoder, androidHeaders, storeEndpoint
	case ANDROID && ERROR:
		return body, updateTimestamp, pyEncoder, androidHeaders, storeEndpoint
	case JAVASCRIPT && TRANSACTION:
		return body, updateTimestamps, jsEncoder, jsHeaders, storeEndpoint
	case JAVASCRIPT && ERROR:
		return body, updateTimestamp, jsEncoder, jsHeaders, storeEndpoint
	case PYTHON && TRANSACTION:
		return body, updateTimestamps, pyEncoder, pyHeaders, storeEndpoint
	case PYTHON && ERROR:
		return body, updateTimestamp, pyEncoder, pyHeaders, storeEndpoint
	}
	return body, updateTimestamps, jsEncoder, jsHeaders, storeEndpoint
}

func buildRequest(requestBody []byte, headerKeys []string, eventHeaders map[string]string, storeEndpoint string) *http.Request {
	request, errNewRequest := http.NewRequest("POST", storeEndpoint, bytes.NewReader(requestBody)) // &buf
	if errNewRequest != nil {
		log.Fatalln(errNewRequest)
	}

	headerInterface := eventHeaders

	for _, v := range headerKeys {
		if headerInterface[v] == "" {			
			fmt.Print("PASS")
		} else {
			request.Header.Set(v, headerInterface[v])
		}
	}
	return request
}

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
	month := date.Month()
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
	release := fmt.Sprint(int(month), ".", week)
	body["release"] = release
	fmt.Println("> release", body["release"])
	return body
}

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
	tags["undertaker"] = "first"

	if _, ok := body["transaction"]; ok {
		if (body["transaction"].(string) == "http://localhost:5000/" || strings.Contains(body["transaction"].(string), "wcap")) {
			body["transaction"] = "http://toolstoredemo.com/"
			request := body["request"].(map[string]interface{})
			request["url"] = "http://toolstoredemo.com/"
			tags["url"] = "http://toolstoredemo.com/"
			spans := body["spans"].([]interface{})
			for _, v1 := range spans {
				v := v1.(map[string]interface{})
				if strings.Contains(v["description"].(string), "wcap") {
					v["description"] = "GET http://toolstoredemo.com/tools"
				}
			}
		}
	}
}
