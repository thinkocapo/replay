package main

import (
	_ "github.com/mattn/go-sqlite3"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"

	// "github.com/buger/jsonparser"
	// "strconv"
	
	"strings"
	"time"
)

func main() {

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
		var body []byte
		var headers string
		rows.Scan(&id, &name, &_type, &body, &headers)
		
		// fmt.Println("LENGTH", len(rows))
		// fmt.Println(headers)

		// only for body (Gzipped)
		r, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			fmt.Println(err)
		}
		body, err = ioutil.ReadAll(r)
		if err != nil {
			fmt.Println(err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			panic(err)
		}

		// EVENT ID
		// is the json
		fmt.Println(data["event_id"])

		// need uuid4
		var _uuid = uuid.New().String()

		_uuid = strings.ReplaceAll(_uuid, "-", "") 
		data["event_id"] = _uuid

		fmt.Println(data["event_id"])

		// TIMESTAMP in format 2020-04-18T23:31:48.710876Z
		currentTime := time.Now()
		former := currentTime.Format("2006-01-02") + "T" + currentTime.Format("15:04:05")

		timestamp := data["timestamp"].(string)
		latter := timestamp[19:]
		
		data["timestamp"] = former + latter
		fmt.Println(data["timestamp"])

		// TODO
		// CONVERT 'data' from go object / json into (encoded) utf8 bytes w/ gzip

		// HTTP EXAMPLE - works...
		resp, err := http.Get("http://example.com/")
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		responseBodyBytes, err := ioutil.ReadAll(resp.Body)
		fmt.Println(string(responseBodyBytes))
		
		// TODO
		// SEND to Sentry via HTTP
			// URL string with sentry key
			// w/ headers, payload
		
		// POST
		// resp, err := http.Post("http://example.com/upload", "image/jpeg", &buf)
		
		// req, err := http.NewRequest("GET", "http://example.com", nil)
		// // ...
		// req.Header.Add("If-None-Match", `W/"wyzzy"`)
		// resp, err := client.Do(req)
		
		// need this? because not reading a bytes object from database anymore
		// decodeBody(body, proto.Header(payload, HTTP_CONTENT_ENCODING))
	}

	rows.Close()

	// TODO get '1' when there's multiple rows available
}


// type Event struct {
// 	id         int
// 	name   string
// 	// type    string
// 	payload []byte
// 	headers []byte
// }

// event_id, err := jsonparser.GetString(body, "event_id")
// if err != nil {
// 	fmt.Println(err)
// }
// fmt.Println("id", strconv.Itoa(id), event_id)