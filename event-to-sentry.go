package main

import (
	_ "github.com/mattn/go-sqlite3"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"

	// "github.com/buger/jsonparser"
	// "strconv"
	
	"strings"
	"time"
)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Messages
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

		var bodyGoInterface map[string]interface{}
		if err := json.Unmarshal(body, &bodyGoInterface); err != nil {
			panic(err)
		}

		// EVENT ID
		// is the json
		fmt.Println(bodyGoInterface["event_id"])

		// need uuid4
		var _uuid = uuid.New().String()

		_uuid = strings.ReplaceAll(_uuid, "-", "") 
		bodyGoInterface["event_id"] = _uuid

		fmt.Println(bodyGoInterface["event_id"])

		// TIMESTAMP in format 2020-04-18T23:31:48.710876Z
		currentTime := time.Now()
		former := currentTime.Format("2006-01-02") + "T" + currentTime.Format("15:04:05")

		timestamp := bodyGoInterface["timestamp"].(string)
		latter := timestamp[19:]
		
		bodyGoInterface["timestamp"] = former + latter
		fmt.Println(bodyGoInterface["timestamp"])

		// TODO.....
		SENTRY_URL := "http://localhost:9000/api/2/store/?sentry_URL_key=09aa0d909232457a8a6dfff118bac658&sentry_version=7"
		postBody, errPostBody := json.Marshal(bodyGoInterface) // CONVERT 'data' from go object / json into (encoded) utf8 bytes w/ gzip?
		postBodyEncoded := encode(postBody) // ioutil writer and gzip?
		buffer := bytes.NewBuffer(postBody)

		// *[]byte does not implement io.Reader (missing Read method)
		// resp, err := http.Post(SENTRY, "image/jpeg", &postBodyEncoded)
		
		reqObject, errNewRequest := http.NewRequest("POST", SENTRY_URL, buffer)
		if errNewRequest != nil { log.Fatalln(errNewRequest) }

		client := &http.Client{
			// CheckRedirect: redirectPolicyFunc,
		}
		// TODO - add abunch of these
		reqObject.Header.Add("If-None-Match", `W/"wyzzy"`)
		// ...
		resp1, err1 := client.Do(reqObject)
		
		// need this? because not reading a bytes object from database anymore
		// decodeBody(body, proto.Header(payload, HTTP_CONTENT_ENCODING))

		// might need a Transport for compression...
	}

	rows.Close()

	
}

func encode(buf []byte) []byte {
	dst := make([]byte, len(buf)*2+1)
	hex.Encode(dst, buf)
	dst[len(dst)-1] = '\n'

	return dst
}



// HTTP EXAMPLE - works...
// resp, err := http.Get("http://example.com/")
// if err != nil {
	// 	fmt.Println(err)
	// }
	// defer resp.Body.Close()
	// responseBodyBytes, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(responseBodyBytes))


// https://golang.org/pkg/net/http/#Client
// https://medium.com/@masnun/making-http-requests-in-golang-dd123379efe7
// resp, errPost := http.Post(SENTRY_URL, "image/jpeg", buffy)

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