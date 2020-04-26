package main

import (
	_ "github.com/mattn/go-sqlite3"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"os"
	"net/http"
	// "github.com/buger/jsonparser"
	"strings"
	"time"
)

var all = flag.Bool("all", false, "send all events or 1 event from database")

var httpClient = &http.Client{
	// CheckRedirect: redirectPolicyFunc,
}


func main() {
	flag.Parse()
	fmt.Println("> --all", *all)

	if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
	}
	// TODO _ is for 'exists' could use in 'func init' to make sure it's there
	DSN, _ := os.LookupEnv("DSN")

	// DSN := os.Getenv("DSN")
	fmt.Println("> DSN", DSN)
	KEY := strings.Split(DSN, "@")[0][7:]
	SENTRY_URL := strings.Join([]string{"http://localhost:9000/api/2/store/?sentry_key=",KEY,"&sentry_version=7"}, "")
	fmt.Println("> SENTRY_URL", SENTRY_URL)

	db, _ := sql.Open("sqlite3", "sqlite.db")
	rows, err := db.Query("SELECT * FROM events")
	if err != nil {
		fmt.Println("We got Error", err)
	}
	for rows.Next() {
		var id int
		var name, _type, headers string
		var bodyBytesCompressed []byte
		
		// TODO	- Struct?
		rows.Scan(&id, &name, &_type, &bodyBytesCompressed, &headers)

		bodyBytes := decodeGzip(bodyBytesCompressed)
		bodyInterface := unmarshalJSON(bodyBytes)

		bodyInterface = replaceEventId(bodyInterface)
		bodyInterface = replaceTimestamp(bodyInterface)
		
		bodyBytesPost := marshalJSON(bodyInterface)
		buf := encodeGzip(bodyBytesPost)

		request, errNewRequest := http.NewRequest("POST", SENTRY_URL, &buf)
		if errNewRequest != nil { log.Fatalln(errNewRequest) }

		headerInterface := unmarshalJSON([]byte(headers))

		for _, v := range [6]string{"Host", "Accept-Encoding","Content-Length","Content-Encoding","Content-Type","User-Agent"} {
			request.Header.Set(v, headerInterface[v].(string))
		}

		response, requestErr := httpClient.Do(request)
		if requestErr != nil { fmt.Println(requestErr) }

		responseData, responseDataErr := ioutil.ReadAll(response.Body)
		if responseDataErr != nil { log.Fatal(responseDataErr) }

		fmt.Printf("> event %v\n", string(responseData))

		if !*all {
			rows.Close()
		}
	}
	rows.Close()
}

func decodeGzip(bodyBytes []byte) []byte {
	bodyReader, err := gzip.NewReader(bytes.NewReader(bodyBytes))
	if err != nil {
		fmt.Println(err)
	}
	bodyBytes, err = ioutil.ReadAll(bodyReader)
	if err != nil {
		fmt.Println(err)
	}
	return bodyBytes
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
	fmt.Println("before",bodyInterface["event_id"])
	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "") 
	bodyInterface["event_id"] = uuid4
	fmt.Println("after ",bodyInterface["event_id"])
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