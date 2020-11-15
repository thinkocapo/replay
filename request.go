package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Request struct {
	Payload       []byte
	StoreEndpoint string
}

func NewRequest(event EventJson) *Request {
	r := new(Request)
	if event.Kind == "error" {
		// fmt.Println("\n> timestamp AFTER 2", event.Error.Timestamp)
		r.StoreEndpoint = dsnToStoreEndpoint(projectDSNs, event.Error.Platform)

		bodyBytes, errBodyBytes := json.Marshal(event.Error)
		if errBodyBytes != nil {
			fmt.Println(errBodyBytes)
		}
		r.Payload = bodyBytes
	}
	if event.Kind == "transaction" {
		r.StoreEndpoint = dsnToStoreEndpoint(projectDSNs, event.Transaction.Platform)

		bodyBytes, errBodyBytes := json.Marshal(event.Transaction)
		if errBodyBytes != nil {
			fmt.Println(errBodyBytes)
		}
		r.Payload = bodyBytes
	}
	// TODO `r.Payload = bodyBytes` here
	// TODO check if either Payload or StoreEndpoint are nil
	// log.Fatal("unrecognized event.Kind", event.Kind)
	return r
}

func (r Request) send() bool {

	// DEPRECATING (moved this to NewRequest)
	// bodyBytes, errBodyBytes := json.Marshal(r.errorPayload)
	// bodyBytes, errBodyBytes := json.Marshal(r.Error)
	// if errBodyBytes != nil {
	// 	fmt.Println(errBodyBytes)
	// }

	request, errNewRequest := http.NewRequest("POST", r.StoreEndpoint, bytes.NewReader(r.Payload)) // &buf
	if errNewRequest != nil {
		log.Fatalln(errNewRequest)
	}

	request.Header.Set("content-type", "application/json")

	fmt.Printf("\n> storeEndpoint %v\n", r.StoreEndpoint)
	// fmt.Printf("\n> errorPayload %+v \n", r.Payload)

	// fmt.Print("XXXXXX Value of ignore XXXXXXX", ignore)
	if *ignore == false {
		response, requestErr := httpClient.Do(request)
		if requestErr != nil {
			log.Fatal(requestErr)
		}
		responseData, responseDataErr := ioutil.ReadAll(response.Body)
		if responseDataErr != nil {
			log.Fatal(responseDataErr)
		}
		fmt.Printf("> KIND|RESPONSE: %s \n", string(responseData))
	} else {
		fmt.Print("> event IGNORED \n")
	}
	return true
}

// DEPRECATING...
func sendRequests(requests []Request) bool {
	for _, request := range requests {
		// if !ignore {
		request.send()
		// } else {
		// fmt.Printf("> %s event IGNORED \n", request.storeEndpoint)
		// }
		time.Sleep(750 * time.Millisecond)
	}
	return true
}

// request = Request{
// 	EventJson:     event,
// 	storeEndpoint: dsnToStoreEndpoint(projectDSNs, event.Error.Platform), // TODO move to Request constructor
// })

// request = Request{
// 	EventJson:     event,
// 	storeEndpoint: dsnToStoreEndpoint(projectDSNs, event.Transaction.Platform), // TODO move to Request constructor
// })

// TODO 12;33p since both events are appearing, then don't need this:
// request.Header.Set("x-sentry-auth", os.Getenv("SENTRY_AUTH_KEY"))
// fmt.Printf("\n> x-sentry-auth %v\n", os.Getenv("SENTRY_AUTH_KEY"))
