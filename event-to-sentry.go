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
	"io/ioutil"
	"log"
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
	fmt.Println("all", *all)

	db, _ := sql.Open("sqlite3", "sqlite.db")
	rows, err := db.Query("SELECT * FROM events")
	if err != nil {
		fmt.Println("We got Error", err)
	}
	for rows.Next() {
		var id int
		var name string
		var _type string
		var bodyBytesCompressed []byte
		var headers string
		
		rows.Scan(&id, &name, &_type, &bodyBytesCompressed, &headers)

		bodyBytes := decodeGzip(bodyBytesCompressed)

		var bodyInterface map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyInterface); err != nil {
			panic(err)
		}

		fmt.Println("before",bodyInterface["event_id"])
		var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "") 
		bodyInterface["event_id"] = uuid4
		fmt.Println("after ",bodyInterface["event_id"])
		
		fmt.Println("before",bodyInterface["timestamp"])
		timestamp := time.Now()
		oldTimestamp := bodyInterface["timestamp"].(string)
		newTimestamp := timestamp.Format("2006-01-02") + "T" + timestamp.Format("15:04:05")
		bodyInterface["timestamp"] = newTimestamp + oldTimestamp[19:]
		fmt.Println("after ",bodyInterface["timestamp"])
		
		postBody, errPostBody := json.Marshal(bodyInterface) 
		if errPostBody != nil { fmt.Println(errPostBody)}
		buf := encodeGzip(postBody)


		SENTRY_URL := "http://localhost:9000/api/2/store/?sentry_key=09aa0d909232457a8a6dfff118bac658&sentry_version=7"
		request, errNewRequest := http.NewRequest("POST", SENTRY_URL, &buf)
		if errNewRequest != nil { log.Fatalln(errNewRequest) }

		var headerInterface map[string]interface{}
		if err := json.Unmarshal([]byte(headers), &headerInterface); err != nil {
			panic(err)
		}
		headerKeys := [6]string{"Host", "Accept-Encoding","Content-Length","Content-Encoding","Content-Type","User-Agent"}
		for i:=0; i < len(headerKeys); i++ {
			key := headerKeys[i]
			request.Header.Set(key, headerInterface[key].(string))
		}

		response, requestErr := httpClient.Do(request)
		if requestErr != nil { fmt.Println(requestErr) }

		responseData, responseDataErr := ioutil.ReadAll(response.Body)
		if responseDataErr != nil { log.Fatal(responseDataErr) }

		fmt.Println(string(responseData))

		if !*all {
			rows.Close()
		}
	}
	rows.Close()
}

// decode gzip compression
func decodeGzip(bodyBytes []byte) []byte {
	bodyReader, err := gzip.NewReader(bytes.NewReader(bodyBytes)) // only for body (Gzipped)
	if err != nil {
		fmt.Println(err)
	}
	bodyBytes, err = ioutil.ReadAll(bodyReader)
	if err != nil {
		fmt.Println(err)
	}
	return bodyBytes
}
// encode gzip compression
func encodeGzip(b []byte) bytes.Buffer {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(b)
	w.Close()
	// return buf.Bytes()
	return buf
}
