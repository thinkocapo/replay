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

var sendOne = flag.Bool("all", true, "send all events or 1 event from database")

func main() {
	flag.Parse()
	fmt.Println("FLAG", *sendOne)

	db, _ := sql.Open("sqlite3", "sqlite.db")
	fmt.Println("Let's connect Sqlite", db)

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

		// check go sdk for how/where (class) headers object is managed
		headersBytes := []byte(headers)
		var headerInterface map[string]interface{}
		if err := json.Unmarshal(headersBytes, &headerInterface); err != nil {
			panic(err)
		}

		bodyBytes := decodeGzip(bodyBytesCompressed)

		var bodyInterface map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyInterface); err != nil {
			panic(err)
		}

		fmt.Println(bodyInterface["event_id"])
		var uuid4 = uuid.New().String()
		uuid4 = strings.ReplaceAll(uuid4, "-", "") 
		bodyInterface["event_id"] = uuid4
		fmt.Println(bodyInterface["event_id"])
		
		currentTime := time.Now()
		former := currentTime.Format("2006-01-02") + "T" + currentTime.Format("15:04:05")
		timestamp := bodyInterface["timestamp"].(string)
		latter := timestamp[19:]
		bodyInterface["timestamp"] = former + latter
		fmt.Println(bodyInterface["timestamp"])
		
		postBody, errPostBody := json.Marshal(bodyInterface) 
		if errPostBody != nil { fmt.Println(errPostBody)}

		buf := encodeGzip(postBody)

		SENTRY_URL := "http://localhost:9000/api/2/store/?sentry_key=09aa0d909232457a8a6dfff118bac658&sentry_version=7"
		request, errNewRequest := http.NewRequest("POST", SENTRY_URL, &buf)
		if errNewRequest != nil { log.Fatalln(errNewRequest) }
		
		headerKeys := [6]string{"Host", "Accept-Encoding","Content-Length","Content-Encoding","Content-Type","User-Agent"}
		for i:=0; i < len(headerKeys); i++ {
			key := headerKeys[i]
			request.Header.Set(key, headerInterface[key].(string))
		}

		client := &http.Client{
			// CheckRedirect: redirectPolicyFunc,
		}

		httpResponse, httpRequestError := client.Do(request)
		if httpRequestError != nil { 
			fmt.Println("ERRRRORRRRRR")
			fmt.Println(httpRequestError)
		}

		fmt.Println("\n************* RESPONSE *********** \n")
		fmt.Println(httpResponse)

		if *sendOne {
			fmt.Println("ONLY ONCE BUDDY\n")
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
