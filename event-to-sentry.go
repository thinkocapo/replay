package main

import (
	_ "github.com/mattn/go-sqlite3"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/google/uuid"
	"io/ioutil"
	"strconv"
	"strings"
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
		
		event_id, err := jsonparser.GetString(body, "event_id")

		if err != nil {
			fmt.Println(err)
		}
		
		fmt.Println("id", strconv.Itoa(id), event_id)

		var dat map[string]interface{}
		if err := json.Unmarshal(body, &dat); err != nil {
			panic(err)
		}
		// is the json
		// fmt.Println(dat)
		fmt.Println(dat["event_id"])

		// need uuid4
		var _uuid = uuid.New().String()

		_uuid = strings.ReplaceAll(_uuid, "-", "") 
		dat["event_id"] = _uuid

		fmt.Println(dat["event_id"])


	}

	rows.Close()

	// TODO - Send Event to Sentry Instance
	// 1. get '1' when there's multiple rows available
	// 2. jsonparser.GetString(body, "event_id") to make new event_id + timestamp
	// 3. somehow encode+gzip again
	// 4. http request to Sentry,io. 
		// URL string with sentry key
		// w/ headers, payload
}












// type Event struct {
// 	id         int
// 	name   string
// 	// type    string
// 	payload []byte
// 	headers []byte
// }