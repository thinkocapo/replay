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
		var bodyBytes []byte
		var headers string
		rows.Scan(&id, &name, &_type, &bodyBytes, &headers)

		fmt.Println("\n------------- HEADERS ------------")

		headersBytes := []byte(headers)
		var headerInterface map[string]interface{}
		if err := json.Unmarshal(headersBytes, &headerInterface); err != nil {
			panic(err)
		}
		fmt.Println("--------- headerInterface", headerInterface["Host"].(string))
		fmt.Println("--------- headerInterface", headerInterface["Accept-Encoding"].(string))
		fmt.Println("--------- headerInterface", headerInterface["Content-Length"].(string))
		fmt.Println("--------- headerInterface", headerInterface["Content-Encoding"].(string))
		fmt.Println("--------- headerInterface", headerInterface["Content-Type"].(string))
		fmt.Println("--------- headerInterface", headerInterface["User-Agent"].(string))


		// DECODE DATA FROM DB
		bodyReader, err := gzip.NewReader(bytes.NewReader(bodyBytes)) // only for body (Gzipped)
		if err != nil {
			fmt.Println(err)
		}
		bodyBytes, err = ioutil.ReadAll(bodyReader)
		if err != nil {
			fmt.Println(err)
		}

		// UNMARSHAL THE BYTES INTO OBJECT
		var bodyInterface map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyInterface); err != nil {
			panic(err)
		}
		
		// PREPARE THE OBJECT's
		// EVENT ID
		fmt.Println(bodyInterface["event_id"])
		var _uuid = uuid.New().String() // uuid4
		_uuid = strings.ReplaceAll(_uuid, "-", "") 
		bodyInterface["event_id"] = _uuid
		fmt.Println(bodyInterface["event_id"])
		// TIMESTAMP format 2020-04-18T23:31:48.710876Z
		currentTime := time.Now()
		former := currentTime.Format("2006-01-02") + "T" + currentTime.Format("15:04:05")
		timestamp := bodyInterface["timestamp"].(string)
		latter := timestamp[19:]
		bodyInterface["timestamp"] = former + latter
		fmt.Println(bodyInterface["timestamp"])

		// HTTP TO SENTRY
		SENTRY_URL := "http://localhost:9000/api/2/store/?sentry_URL_key=09aa0d909232457a8a6dfff118bac658&sentry_version=7"
		postBody, errPostBody := json.Marshal(bodyInterface) // CONVERT 'data' from go object / json into (encoded) utf8 bytes w/ gzip?
		if errPostBody != nil { fmt.Println(errPostBody)}
		postBodyEncoded := encode(postBody) // ioutil writer and gzip????
		buffer := bytes.NewBuffer(postBodyEncoded) // Note - used to take postBody????
		
		// might be missing zip...
		reqObject, errNewRequest := http.NewRequest("POST", SENTRY_URL, buffer)
		if errNewRequest != nil { log.Fatalln(errNewRequest) }

		client := &http.Client{
			// CheckRedirect: redirectPolicyFunc,
		}
		// TODO - HEADERS.....
		reqObject.Header.Add("Host", headerInterface["Host"].(string))
		reqObject.Header.Add("Accept-Encoding", headerInterface["Accept-Encoding"].(string))
		reqObject.Header.Add("Content-Length", headerInterface["Content-Length"].(string))
		reqObject.Header.Add("Content-Encoding", headerInterface["Content-Encoding"].(string))
		reqObject.Header.Add("Content-Type", headerInterface["Content-Type"].(string))
		reqObject.Header.Add("User-Agent", headerInterface["User-Agent"].(string))

		// fmt.Println("\n************* reqObject \n", reqObject) // Note - what are the nulls on this

		fmt.Println("\n************* client.Do *********** \n")
		httpResponse, httpRequestError := client.Do(reqObject)
		if httpRequestError != nil { fmt.Println(httpRequestError)}
		fmt.Println(httpResponse)
		// need this? because not reading a bytes object from database anymore
		// decodeBody(body, proto.Header(payload, HTTP_CONTENT_ENCODING))

		// might need a Transport for compression...
		// might need gzipEncoder gzip.NewWriter...buf.Bytes()...
	}

	rows.Close()

	
}

func encode(buf []byte) []byte {
	dst := make([]byte, len(buf)*2+1)
	hex.Encode(dst, buf)
	dst[len(dst)-1] = '\n'

	return dst
}

// *[]byte does not implement io.Reader (missing Read method)
// resp, err := http.Post(SENTRY, "image/jpeg", &postBodyEncoded)

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