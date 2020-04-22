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
	// "io"
	"log"
	"net/http"
	// "github.com/buger/jsonparser"
	// "strconv"
	"strings"
	"time"
	// "compress/zlib"
)

// sentry-go Transport layer https://github.com/getsentry/sentry-go/blob/db5e5daf4334b2c9b5341cfcb3bbd1535b923c18/transport.go
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
		fmt.Println("\nLENGTH1", len(bodyBytes))

		// UNMARSHAL THE BYTES INTO OBJECT, prepare the EventId&Timestamp
		var bodyInterface map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyInterface); err != nil {
			panic(err)
		}
		fmt.Println(bodyInterface["event_id"])
		var _uuid = uuid.New().String() // uuid4
		_uuid = strings.ReplaceAll(_uuid, "-", "") 
		bodyInterface["event_id"] = _uuid
		fmt.Println(bodyInterface["event_id"])
		currentTime := time.Now()
		former := currentTime.Format("2006-01-02") + "T" + currentTime.Format("15:04:05")
		timestamp := bodyInterface["timestamp"].(string)
		latter := timestamp[19:]
		bodyInterface["timestamp"] = former + latter
		fmt.Println(bodyInterface["timestamp"])

		
		postBody, errPostBody := json.Marshal(bodyInterface) // CONVERT 'data' from go object / json into (encoded) utf8 bytes w/ gzip?
		if errPostBody != nil { fmt.Println(errPostBody)}
		
		// TRANSPORT
		var buf bytes.Buffer
		SENTRY_URL := "http://localhost:9000/api/2/store/?sentry_key=09aa0d909232457a8a6dfff118bac658&sentry_version=7"
		w := gzip.NewWriter(&buf)
		w.Write(postBody)
		w.Close()
		// var b bytes.Buffer
		// w1 := zlib.NewWriter(&b)
		// w1.Write([]byte(postBody))
		// // w1.Write([]byte(string(postBody)))
		// w1.Close()

		
		fmt.Println("\nLENGTH2", len(buf.Bytes()))
		
		// REQUEST OBJECT
		// TODO try passing bodyBytes (unmodified)
		// buffer := bytes.NewBuffer(postBody)
		request, errNewRequest := http.NewRequest("POST", SENTRY_URL, &buf)
		if errNewRequest != nil { log.Fatalln(errNewRequest) }
		
		// TODO - HEADERS.....
		request.Header.Set("Host", headerInterface["Host"].(string))
		request.Header.Set("Accept-Encoding", headerInterface["Accept-Encoding"].(string))
		request.Header.Set("Content-Length", headerInterface["Content-Length"].(string))
		// request.Header.Set("Content-Length", "1502")
		request.Header.Set("Content-Encoding", headerInterface["Content-Encoding"].(string))
		request.Header.Set("Content-Type", headerInterface["Content-Type"].(string))
		request.Header.Set("User-Agent", headerInterface["User-Agent"].(string))
		
		// fmt.Println("\n************* request \n", request)
		
		client := &http.Client{
			// CheckRedirect: redirectPolicyFunc,
		}
		// EXECUTE HTTP REQUEST
		httpResponse, httpRequestError := client.Do(request)
		if httpRequestError != nil { 
			fmt.Println("ERRRRORRRRRR")
			fmt.Println(httpRequestError)
		}


		fmt.Println("\n************* RESPONSE *********** \n")
		fmt.Println(httpResponse)

		// request, _ := http.NewRequest(
		// 	http.MethodPost,
		// 	"http://localhost:9000/api/2/store/?sentry_key=09aa0d909232457a8a6dfff118bac658&sentry_version=7",//t.dsn.StoreAPIURL().String(),
		// 	buffer, // bytes.NewBuffer(body)
		// )
		// resp, err := client.Do(request)	
		// fmt.Println("\n************* client.Do Done1 *********** \n", request)
		// fmt.Println("\n************* client.Do Done2 *********** \n", resp)
		// fmt.Println("\n************* client.Do err *********** \n", err)
		// works.....so comment out below, why won't above one appear anywhere??? no errors on headers...
		// need this? because not reading a bytes object from database anymore
		// decodeBody(body, proto.Header(payload, HTTP_CONTENT_ENCODING))
	}
	rows.Close()
}

func encode(buf []byte) []byte {
	dst := make([]byte, len(buf)*2+1)
	hex.Encode(dst, buf)
	dst[len(dst)-1] = '\n'

	return dst
}

func gzipEncoder(b []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(b)
	w.Close()
	return buf.Bytes()
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