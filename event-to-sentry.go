package main

import (
	_ "github.com/mattn/go-sqlite3"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	// "github.com/buger/jsonparser"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	// "strconv"
	"strings"
	"time"
)

var httpClient = &http.Client{}

var (
	all *bool
	id *string
	ignore *bool
	db *sql.DB
	dsn DSN
	SENTRY_URL string 
	exists bool
	projects map[string]*DSN
)

type DSN struct { 
	host string
	rawurl string
	key string
	projectId string
}

func parseDSN(rawurl string) (*DSN) {
	key := strings.Split(rawurl, "@")[0][7:]

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
	if (strings.Contains(rawurl, "ingest.sentry.io")) {
		host = "ingest.sentry.io"
	}
	if (strings.Contains(rawurl, "@localhost:")) {
		host = "localhost:9000"
	}

	fmt.Printf("> DSN { host: %s, projectId: %s }\n", host, projectId)
	return &DSN{
		host,
		rawurl,
		key,
		projectId,
	}
}

// Could make a DSN field called 'storeEndpoint' and use this function there to assign the value, during parseDSN
func (d DSN) storeEndpoint() string {
	var fullurl string
	if (d.host == "ingest.sentry.io") {
		fullurl = fmt.Sprint("https://",d.host,"/api/",d.projectId,"/store/?sentry_key=",d.key,"&sentry_version=7")
	}
	if (d.host == "localhost:9000") {
		fullurl = fmt.Sprint("http://",d.host,"/api/",d.projectId,"/store/?sentry_key=",d.key,"&sentry_version=7")
	}
	return fullurl
}

type Event struct {
	id int
	name, _type string
	headers []byte
	bodyBytes []byte
}
func (e Event) String() string {
	return fmt.Sprintf("\n Event { SqliteId: %d, Platform: %s, Type: %s }\n", e.id, e.name, e._type)
}

func init() {
	if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
	}

	projects = make(map[string]*DSN)
	
	// Must use SAAS for AM Performance Transactions as https://github.com/getsentry/sentry's Release 10.0.0 doesn't include Performance yet
	// projects["javascript"] = parseDSN(os.Getenv("DSN_REACT"))
	// projects["python"] = parseDSN(os.Getenv("DSN_PYTHON"))
	projects["javascript"] = parseDSN(os.Getenv("DSN_JAVASCRIPT_SAAS"))
	projects["python"] = parseDSN(os.Getenv("DSN_PYTHON_SAAS"))
	projects["node"] = parseDSN(os.Getenv("DSN_EXPRESS_SAAS"))
	projects["go"] = parseDSN(os.Getenv("DSN_GO_SAAS"))
	projects["ruby"] = parseDSN(os.Getenv("DSN_RUBY_SAAS"))

	all = flag.Bool("all", false, "send all events or 1 event from database")
	id = flag.String("id", "", "id of event in sqlite database")
	ignore = flag.Bool("i", false, "ignore sending the event to Sentry.io")

	flag.Parse()

	db, _ = sql.Open("sqlite3", os.Getenv("SQLITE"))
}

func main() {
	defer db.Close()
	
	query := ""
	if (*id == "") {
		query = "SELECT * FROM events ORDER BY id DESC"
	} else {
		query = strings.ReplaceAll("SELECT * FROM events WHERE id=?", "?", *id)
	}

	rows, err := db.Query(query)
	
	if err != nil {
		fmt.Println("Failed to load rows", err)
	}
	for rows.Next() {
		var event Event
		rows.Scan(&event.id, &event.name, &event._type, &event.bodyBytes, &event.headers)
		fmt.Println(event)

		if (event.name == "javascript") {
			javascript(event)
		}
		if (event.name == "python") {
			python(event)
		}

		if !*all {
			rows.Close()
		}

		time.Sleep(450 * time.Millisecond)
	}
	rows.Close()
}

func javascript(event Event) {
	fmt.Sprintf("> JAVASCRIPT %v %v", event.name, event._type)
	
	bodyInterface := unmarshalJSON(event.bodyBytes)
	bodyInterface = replaceEventId(bodyInterface)

	if (event._type == "error") {
		bodyInterface = updateTimestamp(bodyInterface)
	}
	if (event._type == "transaction") {
		bodyInterface = updateTimestamps(bodyInterface, "javascript")
		if (bodyInterface["transaction"] == "http://localhost:5000/") {
			bodyInterface["transaction"] = "http://toolstoredemo.com/"

			request := bodyInterface["request"].(map[string]interface{})
			request["url"] = "http://toolstoredemo.com/"

			tags := bodyInterface["tags"].(map[string]interface{})
			tags["url"] = "http://toolstoredemo.com/"

			spans := bodyInterface["spans"].([]interface{})
			for _, v1 := range spans {
				v := v1.(map[string]interface{})
				if (v["span_id"] == "9cbd476918dcc9b4") {
					v["description"] = "GET http://toolstoredemo.com/tools"
				}
			}
		}
	}

	undertake(bodyInterface)

	bodyBytesPost := marshalJSON(bodyInterface)
	
	SENTRY_URL = projects["javascript"].storeEndpoint()
	// fmt.Printf("> storeEndpoint %v", SENTRY_URL)

	request, errNewRequest := http.NewRequest("POST", SENTRY_URL, bytes.NewReader(bodyBytesPost))
	if errNewRequest != nil { log.Fatalln(errNewRequest) }
	
	headerInterface := unmarshalJSON(event.headers)
	for _, v := range [4]string{"Accept-Encoding","Content-Length","Content-Type","User-Agent"} {
		request.Header.Set(v, headerInterface[v].(string))
	}
	
	if !*ignore {
		response, requestErr := httpClient.Do(request)
		if requestErr != nil { fmt.Println(requestErr) }
	
		responseData, responseDataErr := ioutil.ReadAll(response.Body)
		if responseDataErr != nil { log.Fatal(responseDataErr) }
	
		fmt.Printf("\n> javascript event response\n", string(responseData))
	} else {
		fmt.Printf("\n> javascript event IGNORED\n")
	}
}

func python(event Event) {
	fmt.Sprintf("> PYTHON %v %v", event.name, event._type)
	// bodyBytes := decodeGzip(bodyBytesCompressed)
	bodyInterface := unmarshalJSON(event.bodyBytes)
	bodyInterface = replaceEventId(bodyInterface)

	if (event._type == "error") {
		bodyInterface = updateTimestamp(bodyInterface)
	}
	if (event._type == "transaction") {
		// bodyInterface = updateTimestamps3(bodyInterface, decimal.NewFromString)
		bodyInterface = updateTimestamps(bodyInterface, "python")
	}

	undertake(bodyInterface)
	
	bodyBytesPost := marshalJSON(bodyInterface)
	buf := encodeGzip(bodyBytesPost)
	
	// TODO control the project from cli, which you want to send to
	SENTRY_URL = projects["python"].storeEndpoint()
	fmt.Printf("> storeEndpoint %v", SENTRY_URL)

	request, errNewRequest := http.NewRequest("POST", SENTRY_URL, &buf)
	if errNewRequest != nil { log.Fatalln(errNewRequest) }

	headerInterface := unmarshalJSON(event.headers)

	// Including X-Sentry-Auth causes, "multiple authorization payloads requested". Why was it being used at one point here? Was it needed for JS errors? It's not used for transactions
	for _, v := range [5]string{"Accept-Encoding","Content-Length","Content-Encoding","Content-Type","User-Agent"} {
		request.Header.Set(v, headerInterface[v].(string))
	}

	if !*ignore {
		response, requestErr := httpClient.Do(request)
		if requestErr != nil { fmt.Println(requestErr) }

		responseData, responseDataErr := ioutil.ReadAll(response.Body)
		if responseDataErr != nil { log.Fatal(responseDataErr) }

		fmt.Printf("\n> python event response: %v\n", string(responseData))
	} else {
		fmt.Printf("\n> python event IGNORED\n")
	}
	
}

// used for ERRORS
// js timestamps https://github.com/getsentry/sentry-javascript/pull/2575
// "1590946750" but as of 06/07/2020 the 'timestamp' property comes in as <nil>. do not need to set the extra decimals
// "2020-05-31T23:55:11.807534Z" for python
// new timestamp format is same for js/python even though was different format on the way in
func updateTimestamp(bodyInterface map[string]interface{}) map[string]interface{} {
	fmt.Println("> Error timestamp before", bodyInterface["timestamp"])
	bodyInterface["timestamp"] = time.Now().Unix() 
	fmt.Println("> Error timestamp after ", bodyInterface["timestamp"])
	return bodyInterface
}

// used for TRANSACTIONS
// start/end here is same as the sdk's start_timestamp/timestamp, and start_timestamp is only present in transactions
// For future reference, data.contexts.trace.span_id is the Parent Trace and at one point I thoguht I saw data.entries with spans. Disregarding it for now.
// Subtraction arithmetic needed on the decimals via Floats, so avoid Int's
// Better to put as Float64 before serialization. also keep to 7 decimal places as the range sent by sdk's is 4 to 7
func updateTimestamps(data map[string]interface{}, platform string) map[string]interface{} {
	// fmt.Printf("\n> both updateTimestamps PARENT start_timestamp before %v (%T) \n", data["start_timestamp"], data["start_timestamp"])
	// fmt.Printf("> both updateTimestamps PARENT       timestamp before %v (%T)", data["timestamp"], data["timestamp"])
	
	var parentStartTimestamp, parentEndTimestamp decimal.Decimal
	// PYTHON timestamp format is 2020-06-06T04:54:56.636664Z RFC3339Nano
	if (platform == "python") {	
		parentStart, _ := time.Parse(time.RFC3339Nano, data["start_timestamp"].(string)) // integer?
		parentEnd, _ := time.Parse(time.RFC3339Nano, data["timestamp"].(string))
		parentStartTime := fmt.Sprint(parentStart.UnixNano())
		parentEndTime := fmt.Sprint(parentEnd.UnixNano())
		parentStartTimestamp, _ = decimal.NewFromString(parentStartTime[:10] + "." + parentStartTime[10:])
		parentEndTimestamp, _ = decimal.NewFromString(parentEndTime[:10] + "." + parentEndTime[10:])
	}
	// JAVASCRIPT timestamp format is 1591419091.4805 to 1591419092.000035
	if (platform == "javascript") {
		// in sqlite it was float64, not a string. or rather, Go is making it a float64 upon reading from db? not sure
		// make into a 'decimal' class type for logging or else it logs as "1.5914674155654302e+09" instead of 1591467415.5654302
		parentStartTimestamp = decimal.NewFromFloat(data["start_timestamp"].(float64))
		parentEndTimestamp = decimal.NewFromFloat(data["timestamp"].(float64))	
	}
	
	// PARENT TRACE
	// 1. Identify the end time / delta is same every time, in Discover, put Discover link here
	// 2. Adjust the parentDifference between .01 and .2 (1% and 20% difference)
	// 3. adjust every span with same 'end timestamp' so they'll always line up with Parent Trace
	// TODO adjust the rate in negative percentage too
	parentDifference := parentEndTimestamp.Sub(parentStartTimestamp)
	fmt.Printf("\n> ****** PARENT DIFFERENCE                 *********", parentDifference)
	rand.Seed(time.Now().UnixNano())
	percentage := 0.01 + rand.Float64() * (0.20 - 0.01)
	fmt.Println("\n> ++++++ percentage", percentage)
	rate := decimal.NewFromFloat(percentage)
	parentDifference = parentDifference.Mul(rate.Add(decimal.NewFromFloat(1)))
	fmt.Printf("\n> ****** PARENT DIFFERENCE updated by rate *********", parentDifference)

	unixTimestampString := fmt.Sprint(time.Now().UnixNano())
	newParentStartTimestamp, _ := decimal.NewFromString(unixTimestampString[:10] + "." + unixTimestampString[10:])
	newParentEndTimestamp := newParentStartTimestamp.Add(parentDifference)

	if (!newParentEndTimestamp.Sub(newParentStartTimestamp).Equal(parentDifference)) {
		fmt.Printf("\nFALSE - parent BOTH", newParentEndTimestamp.Sub(newParentStartTimestamp))
	}

	data["start_timestamp"], _ = newParentStartTimestamp.Round(7).Float64()
	data["timestamp"], _ = newParentEndTimestamp.Round(7).Float64()

	// Could conver back to RFC3339Nano (as that's what the python sdk uses for transactions Python Transactions use) but Floats are working and mirrors what the javascript() function does
	// logging with decimal just so it's more readable and convertible in https://www.epochconverter.com/, because the 'Float' form is like 1.5914674155654302e+09
	fmt.Printf("\n> both updateTimestamps PARENT start_timestamp after %v (%T) \n", decimal.NewFromFloat(data["start_timestamp"].(float64)), data["start_timestamp"])
	fmt.Printf("> both updateTimestamps PARENT       timestamp after %v (%T) \n", decimal.NewFromFloat(data["timestamp"].(float64)), data["timestamp"])

	// SPAN
	// TEST for making sure that the span object was updated by reference
	// firstSpan := data["spans"].([]interface{})[0].(map[string]interface{})
	// fmt.Printf("\n> before ", decimal.NewFromFloat(firstSpan["start_timestamp"].(float64)))
	for _, span := range data["spans"].([]interface{}) {
		sp := span.(map[string]interface{})
		// fmt.Printf("\n> both updatetimestamps SPAN start_timestamp before %v (%T)", sp["start_timestamp"], sp["start_timestamp"])
		// fmt.Printf("\n> both updatetimestamps SPAN       timestamp before %v (%T)\n", sp["timestamp"]	, sp["timestamp"])

		var spanStartTimestamp, spanEndTimestamp decimal.Decimal
		if (platform == "python") {
			spanStart, _ := time.Parse(time.RFC3339Nano, sp["start_timestamp"].(string))
			spanEnd, _ := time.Parse(time.RFC3339Nano, sp["timestamp"].(string))
			spanStartTime := fmt.Sprint(spanStart.UnixNano())
			spanEndTime := fmt.Sprint(spanEnd.UnixNano())
			spanStartTimestamp, _ = decimal.NewFromString(spanStartTime[:10] + "." + spanStartTime[10:])
			spanEndTimestamp, _ = decimal.NewFromString(spanEndTime[:10] + "." + spanEndTime[10:])
		}
		if (platform == "javascript") {
			spanStartTimestamp = decimal.NewFromFloat(sp["start_timestamp"].(float64))
			spanEndTimestamp = decimal.NewFromFloat(sp["timestamp"].(float64))		
		}

		spanDifference := spanEndTimestamp.Sub(spanStartTimestamp)
		spanToParentDifference := spanStartTimestamp.Sub(parentStartTimestamp)
		
		// should use newParentStartTimestamp instead of spanStartTimestamp?
		unixTimestampString := fmt.Sprint(time.Now().UnixNano())
		unixTimestampDecimal, _ := decimal.NewFromString(unixTimestampString[:10] + "." + unixTimestampString[10:])
		newSpanStartTimestamp := unixTimestampDecimal.Add(spanToParentDifference)
		newSpanEndTimestamp := newSpanStartTimestamp.Add(spanDifference)
	
		if (!newSpanEndTimestamp.Sub(newSpanStartTimestamp).Equal(spanDifference)) {
			fmt.Printf("\nFALSE - span BOTH", newSpanEndTimestamp.Sub(newSpanStartTimestamp))
		}

		sp["start_timestamp"], _ = newSpanStartTimestamp.Round(7).Float64()
		sp["timestamp"], _ = newSpanEndTimestamp.Round(7).Float64()

		// logging with decimal just so it's more readable and convertible in https://www.epochconverter.com/, because the 'Float' form is like 1.5914674155654302e+09
		// fmt.Printf("\n> both updatetimestamps SPAN start_timestamp after %v (%T)", decimal.NewFromFloat(sp["start_timestamp"].(float64)), sp["start_timestamp"])
		// fmt.Printf("\n> both updatetimestamps SPAN       timestamp after %v (%T)\n", decimal.NewFromFloat(sp["timestamp"].(float64)), sp["timestamp"])
	}
	// TESt for making sure that the span object was updated by reference. E.g. 1591467416.0387652 should now be 1591476953.491206959
	// fmt.Printf("\n> after ", firstSpan["start_timestamp"])
	return data
}

func replaceEventId(bodyInterface map[string]interface{}) map[string]interface{} {
	if _, ok := bodyInterface["event_id"]; !ok { 
		log.Print("no event_id on object from DB")
	}
	// fmt.Println("> before",bodyInterface["event_id"])
	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "") 
	bodyInterface["event_id"] = uuid4
	fmt.Println("> event_id after",bodyInterface["event_id"])
	return bodyInterface
}

// Python Error Events do not have 'tags' attribute, if no custom tags were set...? "Sometimes there's no tags attribute yet (typically if no custom tags were set, at least for ERr EVents". Transactions come with a few tags by default, by the sdk.
func undertake(bodyInterface map[string]interface{}) {
	if (bodyInterface["tags"] == nil) {
		bodyInterface["tags"] = make(map[string]interface{})
	}
	tags := bodyInterface["tags"].(map[string]interface{})
	tags["undertaker"] = "crontab"

	// Optional - overwrite the platform (make sure matches the DSN's project type)
	// bodyInterface["platform"] = "ruby"
	// Optional - overwrite what the transaction's title will display as in Discover
	// bodyInterface["transaction"] = "eprescription/:id"
}

////////////////////////////  UTILS  /////////////////////////////////////////
func decodeGzip(bodyBytesInput []byte) (bodyBytesOutput []byte) {
	bodyReader, err := gzip.NewReader(bytes.NewReader(bodyBytesInput))
	if err != nil {
		fmt.Println(err)
	}
	bodyBytesOutput, err = ioutil.ReadAll(bodyReader)
	if err != nil {
		fmt.Println(err)
	}
	return
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

//////////////////////////////////////////////////////////////////////////
// example type add func(a int, b int) int
// https://golang.org/pkg/go/types/
// func updateTimestamps3(data map[string]interface{}, platform string, dec func(*decimal.Decimal) decimal.Decimal) map[string]interface{} {
// 	return data
// }