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
	// TODO move `r.Payload = bodyBytes` to down here
	// TODO check if either Payload or StoreEndpoint are nil
	// log.Fatal("unrecognized event.Kind", event.Kind)
	return r
}

func (r Request) send() bool {
	request, errNewRequest := http.NewRequest("POST", r.StoreEndpoint, bytes.NewReader(r.Payload)) // &buf
	if errNewRequest != nil {
		log.Fatalln(errNewRequest)
	}

	request.Header.Set("content-type", "application/json")

	fmt.Printf("\n> storeEndpoint %v\n", r.StoreEndpoint)

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
		request.send()
		time.Sleep(750 * time.Millisecond)
	}
	return true
}
