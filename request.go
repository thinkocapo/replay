package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/getsentry/sentry-go"
)

type Request struct {
	Payload       []byte
	StoreEndpoint string
}

func NewRequest(event Event) *Request {
	r := new(Request)

	var bodyBytes []byte
	var err error
	if event.Kind == ERROR || event.Kind == DEFAULT {
		bodyBytes, err = json.Marshal(event.Error)
	}
	if event.Kind == TRANSACTION {
		bodyBytes, err = json.Marshal(event.Transaction)
	}
	if err != nil {
		sentry.CaptureException(err)
		fmt.Println(err)
	}

	r.Payload = bodyBytes
	r.StoreEndpoint = event.DSN.storeEndpoint()

	if r.StoreEndpoint == "" || r.Payload == nil {
		sentry.CaptureException(errors.New("missing StoreEndpoint or Payload"))
		log.Fatal("missing StoreEndpoint or Payload")
	}
	return r
}

func (r Request) send() bool {
	request, errNewRequest := http.NewRequest("POST", r.StoreEndpoint, bytes.NewReader(r.Payload)) // &buf
	if errNewRequest != nil {
		sentry.CaptureException(errNewRequest)
		log.Fatalln(errNewRequest)
	}

	request.Header.Set("content-type", "application/json")

	fmt.Printf("\n> storeEndpoint %v\n", r.StoreEndpoint)

	if *ignore == false {
		var httpClient = &http.Client{}
		response, requestErr := httpClient.Do(request)
		if requestErr != nil {
			sentry.CaptureException(requestErr)
			log.Fatal(requestErr)
		}
		responseData, responseDataErr := ioutil.ReadAll(response.Body)
		if responseDataErr != nil {
			sentry.CaptureException(responseDataErr)
			log.Fatal(responseDataErr)
		}
		fmt.Printf("> KIND|RESPONSE: %s \n", string(responseData))
	} else {
		fmt.Print("> event IGNORED \n")
	}
	return true
}
