package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Request struct {
	Payload       []byte
	StoreEndpoint string
}

func NewRequest(event Event) *Request {
	r := new(Request)

	var bodyBytes []byte
	var err error
	if event.Kind == ERROR {
		bodyBytes, err = json.Marshal(event.Error)
	}
	if event.Kind == TRANSACTION {
		bodyBytes, err = json.Marshal(event.Transaction)
	}
	if err != nil {
		fmt.Println(err)
	}

	r.Payload = bodyBytes
	r.StoreEndpoint = event.DSN.storeEndpoint()

	if r.StoreEndpoint == "" || r.Payload == nil {
		fmt.Println("something was nil")
	}
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
		var httpClient = &http.Client{}
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
