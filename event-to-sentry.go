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
		// still works if you pass in the "o87286"
		// fullurl = fmt.Sprint("https://o87286.",d.host,"/api/",d.projectId,"/store/?sentry_key=",d.key,"&sentry_version=7")	
		// fullurl = fmt.Sprint("https://",d.host,"/api/",d.projectId,"/store/")
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
	
	// Must use Hosted Sentry for AM Performance Transactions
	// projects["javascript"] = parseDSN(os.Getenv("DSN_REACT"))
	// projects["python"] = parseDSN(os.Getenv("DSN_PYTHON"))
	projects["javascript"] = parseDSN(os.Getenv("DSN_REACT_SAAS"))
	projects["python"] = parseDSN(os.Getenv("DSN_PYTHONTEST_SAAS"))

	all = flag.Bool("all", false, "send all events or 1 event from database")
	id = flag.String("id", "", "id of event in sqlite database")
	flag.Parse()

	db, _ = sql.Open("sqlite3", "am-transactions-sqlite.db")
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
	}
	rows.Close()
}

func javascript(event Event) {
	fmt.Sprintf("> JAVASCRIPT %v %v", event.name, event._type)
	
	bodyInterface := unmarshalJSON(event.bodyBytes)
	bodyInterface = replaceEventId(bodyInterface)

	if (event._type == "error") {
		bodyInterface = updateTimestamp(bodyInterface, "javascript")
	}
	if (event._type == "transaction") {
		bodyInterface = updateTimestamps(bodyInterface, "javascript")
	}

	// undertake()

	bodyBytesPost := marshalJSON(bodyInterface)
	
	SENTRY_URL = projects["javascript"].storeEndpoint()
	fmt.Printf("> storeEndpoint %v", SENTRY_URL)

	request, errNewRequest := http.NewRequest("POST", SENTRY_URL, bytes.NewReader(bodyBytesPost))
	if errNewRequest != nil { log.Fatalln(errNewRequest) }
	
	headerInterface := unmarshalJSON(event.headers)
	for _, v := range [4]string{"Accept-Encoding","Content-Length","Content-Type","User-Agent"} {
		request.Header.Set(v, headerInterface[v].(string))
	}
	
	response, requestErr := httpClient.Do(request)
	if requestErr != nil { fmt.Println(requestErr) }

	responseData, responseDataErr := ioutil.ReadAll(response.Body)
	if responseDataErr != nil { log.Fatal(responseDataErr) }

	// TODO this prints nicely if response is coming from Self-Hosted. Not the case when sending to Hosted sentry
	fmt.Printf("\n> javascript event response\n", string(responseData))
}

func python(event Event) {
	fmt.Sprintf("> PYTHON %v %v", event.name, event._type)
	// bodyBytes := decodeGzip(bodyBytesCompressed)
	bodyInterface := unmarshalJSON(event.bodyBytes)
	bodyInterface = replaceEventId(bodyInterface)

	if (event._type == "error") {
		bodyInterface = updateTimestamp(bodyInterface, "python")
	}
	if (event._type == "transaction") {
		bodyInterface = updateTimestamps(bodyInterface, "python")
	}

	// undertake()
	
	bodyBytesPost := marshalJSON(bodyInterface)
	buf := encodeGzip(bodyBytesPost)
	
	SENTRY_URL = projects["python"].storeEndpoint()
	fmt.Printf("> storeEndpoint %v", SENTRY_URL)

	request, errNewRequest := http.NewRequest("POST", SENTRY_URL, &buf)
	if errNewRequest != nil { log.Fatalln(errNewRequest) }

	headerInterface := unmarshalJSON(event.headers)

	// Including X-Sentry-Auth causes, "multiple authorization payloads requested". Why was it being used at one point here? Was it needed for JS errors? It's not used for transactions
	for _, v := range [5]string{"Accept-Encoding","Content-Length","Content-Encoding","Content-Type","User-Agent"} {
		request.Header.Set(v, headerInterface[v].(string))
	}

	response, requestErr := httpClient.Do(request)
	if requestErr != nil { fmt.Println(requestErr) }

	responseData, responseDataErr := ioutil.ReadAll(response.Body)
	if responseDataErr != nil { log.Fatal(responseDataErr) }

	fmt.Printf("\n> python event response: %v\n", string(responseData))
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

// js timestamps https://github.com/getsentry/sentry-javascript/pull/2575
func updateTimestamp(bodyInterface map[string]interface{}, platform string) map[string]interface{} {
	fmt.Println(" timestamp before", bodyInterface["timestamp"]) // nil for js errors, despite being on latest sdk as of 05/30/2020
	
	// "1590946750"
	// TODO - works? or need the extra decimals (millseconds) at the end
	if (platform == "javascript") {
		bodyInterface["timestamp"] = time.Now().Unix() 
	}

	// "2020-05-31T23:55:11.807534Z"
	if (platform == "python") {
		// is PST, or wherever you're running this from
		timestamp := time.Now()
		// is GMT, so not same as timezone you're running this from
		oldTimestamp := bodyInterface["timestamp"].(string)
		newTimestamp := timestamp.Format("2006-01-02") + "T" + timestamp.Format("15:04:05")
		bodyInterface["timestamp"] = newTimestamp + oldTimestamp[19:]

		// TODO these should match. 'timestamp before' is GTC, appearing as far ahead of PST.
		// timestamp before 2020-06-02T00:09:51.365214Z
		// timestamp after  2020-06-01T17:12:26.365214Z
	}

	fmt.Println("  timestamp after", bodyInterface["timestamp"])
	return bodyInterface
}


// start/end here is same as the sdk's start_timestamp/timestamp, and start_timestamp is only present in transactions
// For future reference, data.contexts.trace.span_id is the Parent Trace and at one point I thoguht I saw data.entries with spans. Disregarding it for now.
// Subtraction arithmetic needed on the decimals via Floats, so avoid Int's
// Better to put as Float64 before serialization. also keep to 7 decimal places as the range sent by sdk's is 4 to 7
func updateTimestamps(data map[string]interface{}, platform string) map[string]interface{} {
	// PYTHON timestamp format is 2020-06-06T04:54:56.636664Z RFC3339Nano
	if (platform == "python") {
		fmt.Printf("\n> py updateTimestamps parent start_timestamp before %v (%T) \n", data["start_timestamp"], data["start_timestamp"])
		fmt.Printf("> py updateTimestamps parent       timestamp before %v (%T)", data["timestamp"], data["timestamp"])
		
		// PARENT TRACE
		// TODO rename as parentStartTime (i.e. object from Go's Time package) and parentStartTimeString, because parentStartTimestamp follows that logically
		t1, _ := time.Parse(time.RFC3339Nano, data["start_timestamp"].(string))
		t2, _ := time.Parse(time.RFC3339Nano, data["timestamp"].(string))
		t1String := fmt.Sprint(t1.UnixNano())
		t2String := fmt.Sprint(t2.UnixNano())
	
		parentStartTimestamp, _ := decimal.NewFromString(t1String[:10] + "." + t1String[10:])
		parentEndTimestamp, _ := decimal.NewFromString(t2String[:10] + "." + t2String[10:])
		parentDifference := parentEndTimestamp.Sub(parentStartTimestamp)

		unixTimestampString := fmt.Sprint(time.Now().UnixNano())
		newParentStartTimestamp, _ := decimal.NewFromString(unixTimestampString[:10] + "." + unixTimestampString[10:])
		newParentEndTimestamp := newParentStartTimestamp.Add(parentDifference)

		if (newParentEndTimestamp.Sub(newParentStartTimestamp).Equal(parentDifference)) {
			fmt.Printf("\nTRUE - parent PYTHON")
		} else {
			fmt.Printf("\nFALSE - parent PYTHON")
			fmt.Print(newParentEndTimestamp.Sub(newParentStartTimestamp))
		}
		data["start_timestamp"], _ = newParentStartTimestamp.Round(7).Float64()
		data["timestamp"], _ = newParentEndTimestamp.Round(7).Float64()

		// Could conver back to RFC3339Nano (as that's what the python sdk uses for transactions Python Transactions use) but Floats are working and mirrors what the javascript() function does
		// logging with decimal just so it's more readable and convertible in https://www.epochconverter.com/, because the 'Float' form is like 1.5914674155654302e+09
		fmt.Printf("\n> py updateTimestamps parent start_timestamp after %v (%T) \n", decimal.NewFromFloat(data["start_timestamp"].(float64)), data["start_timestamp"])
		fmt.Printf("> py updateTimestamps parent       timestamp after %v (%T) \n", decimal.NewFromFloat(data["timestamp"].(float64)), data["timestamp"])

		// SPANS
		for _, span := range data["spans"].([]interface{}) {
			sp := span.(map[string]interface{})
			
			fmt.Printf("\n> py updatetimestamps SPAN start_timestamp before %v (%T)", sp["start_timestamp"], sp["start_timestamp"])
			fmt.Printf("\n> py updatetimestamps SPAN       timestamp before %v (%T)\n", sp["timestamp"]	, sp["timestamp"])
			t1, _ := time.Parse(time.RFC3339Nano, sp["start_timestamp"].(string))
			t2, _ := time.Parse(time.RFC3339Nano, sp["timestamp"].(string))
			t1String := fmt.Sprint(t1.UnixNano())
			t2String := fmt.Sprint(t2.UnixNano())

			spanStartTimestamp, _ := decimal.NewFromString(t1String[:10] + "." + t1String[10:])
			spanEndTimestamp, _ := decimal.NewFromString(t2String[:10] + "." + t2String[10:])
			spanDifference := spanEndTimestamp.Sub(spanStartTimestamp)
			spanToParentDifference := spanStartTimestamp.Sub(parentStartTimestamp)
		
			unixTimestampString := fmt.Sprint(time.Now().UnixNano())
			unixTimestampDecimal, _ := decimal.NewFromString(unixTimestampString[:10] + "." + unixTimestampString[10:])
			newSpanStartTimestamp := unixTimestampDecimal.Add(spanToParentDifference)
			newSpanEndTimestamp := newSpanStartTimestamp.Add(spanDifference)
		
			if (newSpanEndTimestamp.Sub(newSpanStartTimestamp).Equal(spanDifference)) {
				fmt.Printf("TRUE - span PYTHON")
			} else {
				fmt.Printf("\nFALSE - span PYTHON")
				fmt.Print(newSpanEndTimestamp.Sub(newSpanStartTimestamp))
			}
			sp["start_timestamp"], _ = newSpanStartTimestamp.Round(7).Float64()
			sp["timestamp"], _ = newSpanEndTimestamp.Round(7).Float64()

			// logging with decimal just so it's more readable and convertible in https://www.epochconverter.com/, because the 'Float' form is like 1.5914674155654302e+09
			fmt.Printf("\n> py updatetimestamps SPAN start_timestamp after %v (%T)", decimal.NewFromFloat(sp["start_timestamp"].(float64)), sp["start_timestamp"])
			fmt.Printf("\n> py updatetimestamps SPAN       timestamp after %v (%T)\n", decimal.NewFromFloat(sp["timestamp"].(float64)), sp["timestamp"])
		}
	}

	// JAVASCRIPT timestamp format is 1591419091.4805 to 1591419092.000035
	if (platform == "javascript") {
		// PARENT TRACE
		// in sqlite it was float64, not a string. or rather, Go is making it a float64 upon reading from db? not sure
		// make into a 'decimal' class type for logging or else it logs as "1.5914674155654302e+09" instead of 1591467415.5654302
		fmt.Printf("> js updateTimestamps parent start_timestamp before %v (%T) \n", decimal.NewFromFloat(data["start_timestamp"].(float64)), decimal.NewFromFloat(data["start_timestamp"].(float64)))
		fmt.Printf("> js updateTimestamps parent       timestamp before %v (%T) \n", decimal.NewFromFloat(data["timestamp"].(float64)), decimal.NewFromFloat(data["timestamp"].(float64)))
		
		parentStartTimestamp := decimal.NewFromFloat(data["start_timestamp"].(float64))
		parentEndTimestamp := decimal.NewFromFloat(data["timestamp"].(float64))		
		parentDifference := parentEndTimestamp.Sub(parentStartTimestamp)
	
		unixTimestampString := fmt.Sprint(time.Now().UnixNano())
		newParentStartTimestamp, _ := decimal.NewFromString(unixTimestampString[:10] + "." + unixTimestampString[10:])
		newParentEndTimestamp := newParentStartTimestamp.Add(parentDifference)
	
		if (newParentEndTimestamp.Sub(newParentStartTimestamp).Equal(parentDifference)) {
			fmt.Printf("\nTRUE - parent")
		} else {
			fmt.Printf("\nFALSE - parent")
			fmt.Print(newParentEndTimestamp.Sub(newParentStartTimestamp))
		}

		data["start_timestamp"], _ = newParentStartTimestamp.Round(7).Float64()
		data["timestamp"], _ = newParentEndTimestamp.Round(7).Float64()

		fmt.Printf("\n> js updatetimestamps parent start_timestamp after %v (%T)\n", data["start_timestamp"], data["start_timestamp"])
		fmt.Printf("\n> js updatetimestamps parent       timestamp after %v (%T)\n", data["timestamp"], data["timestamp"])

		// TEST for making sure that the span object was updated by reference
		// firstSpan := data["spans"].([]interface{})[0].(map[string]interface{})
		// fmt.Printf("\n> before ", decimal.NewFromFloat(firstSpan["start_timestamp"].(float64)))

		// SPANS
		for _, span := range data["spans"].([]interface{}) {
			sp := span.(map[string]interface{})
			
			fmt.Printf("\n> js updatetimestamps SPAN start_timestamp before %v (%T)", decimal.NewFromFloat(sp["start_timestamp"].(float64)), decimal.NewFromFloat(sp["start_timestamp"].(float64)))
			fmt.Printf("\n> js updatetimestamps SPAN       timestamp before %v (%T)\n", decimal.NewFromFloat(sp["timestamp"].(float64))	, decimal.NewFromFloat(sp["timestamp"].(float64)))
			
			spanStartTimestamp := decimal.NewFromFloat(sp["start_timestamp"].(float64))
			spanEndTimestamp := decimal.NewFromFloat(sp["timestamp"].(float64))		
			spanDifference := spanEndTimestamp.Sub(spanStartTimestamp)
			spanToParentDifference := spanStartTimestamp.Sub(parentStartTimestamp)
		
			unixTimestampString := fmt.Sprint(time.Now().UnixNano())
			unixTimestampDecimal, _ := decimal.NewFromString(unixTimestampString[:10] + "." + unixTimestampString[10:])
			newSpanStartTimestamp := unixTimestampDecimal.Add(spanToParentDifference)
			newSpanEndTimestamp := newSpanStartTimestamp.Add(spanDifference)
		
			if (newSpanEndTimestamp.Sub(newSpanStartTimestamp).Equal(spanDifference)) {
				fmt.Printf("TRUE - span")
			} else {
				fmt.Printf("\nFALSE - span")
				fmt.Print(newSpanEndTimestamp.Sub(newSpanStartTimestamp))
			}

			// is okay that this is an instance of the 'decimal' package and no longer Float64? 
			sp["start_timestamp"], _ = newSpanStartTimestamp.Round(7).Float64()
			sp["timestamp"], _ = newSpanEndTimestamp.Round(7).Float64()

			fmt.Printf("\n> js updatetimestamps SPAN start_timestamp after %v (%T)", sp["start_timestamp"], sp["start_timestamp"])
			fmt.Printf("\n> js updatetimestamps SPAN       timestamp after %v (%T)\n", sp["timestamp"], sp["timestamp"])
		}

		// TEST for making sure that the span object was updated by reference. E.g. 1591467416.0387652 should now be 1591476953.491206959
		// fmt.Printf("\n> after ", firstSpan["start_timestamp"])
	} 
	return data
}

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

// TODO - test, does this update by reference? is this how to return nil?
func undertake(bodyInterface map[string]interface{}) {
	tags := bodyInterface["tags"].(map[string]interface{})
	tags["undertaker"] = "is_here"
}