package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// var httpClient = &http.Client{}

var (
// all         *bool
// id          *string
// ignore      *bool
// database    string
// db          *string
// js          *string
// py          *string
// dsn         DSN
// SENTRY_URL  string
)

// TODO maybe?
type Theencoder interface {
	encodeit() []byte
}

type Transport struct {
	kind          string
	platform      string
	eventHeaders  map[string]string
	storeEndpoint string

	bodyError   map[string]interface{}
	bodyEncoder BodyEncoder

	// TODO
	encoded []byte

	envelopeItems   []interface{}
	envelopeEncoder EnvelopeEncoder // TODO Type or a function here?

	// NO, but maybe for 'requests'
	// envelopeItems []interface{}
	// bodyErrors []map[string]
}

// TODO many need....
// func (t Transport) envelopeEncoder() []byte {
// 	return t.
// }

func encodeAndSendEvents(requests []Transport, ignore bool) {
	fmt.Println("\n> encodeAndSendEvents ...............")
	for _, transport := range requests {
		// var encoded []byte
		if transport.kind == "transaction" {
			transport.encoded = transport.envelopeEncoder(transport.envelopeItems)
		}
		if transport.kind == "error" {
			transport.encoded = transport.bodyEncoder(transport.bodyError)
		}
		request := buildRequest(transport.encoded, transport.eventHeaders, transport.storeEndpoint)

		// TODO - the ignore flag...
		if !ignore {
			response, requestErr := httpClient.Do(request)
			if requestErr != nil {
				log.Fatal(requestErr)
			}
			responseData, responseDataErr := ioutil.ReadAll(response.Body)
			if responseDataErr != nil {
				log.Fatal(responseDataErr)
			}
			fmt.Printf("> KIND|RESPONSE: %s %s\n", transport.kind, string(responseData))
		} else {
			fmt.Printf("> %s event IGNORED \n", transport.kind)
		}

		time.Sleep(1000 * time.Millisecond)
	}
}

func buildRequest(requestBody []byte, eventHeaders map[string]string, storeEndpoint string) *http.Request {
	fmt.Printf("> storeEndpoint %v \n", storeEndpoint)
	if requestBody == nil {
		log.Fatalln("buildRequest missing requestBody")
	}
	if eventHeaders == nil {
		log.Fatalln("buildRequest missing eventHeaders")
	}
	if storeEndpoint == "" {
		log.Fatalln("buildRequest missing storeEndpoint")
	}

	request, errNewRequest := http.NewRequest("POST", storeEndpoint, bytes.NewReader(requestBody)) // &buf
	if errNewRequest != nil {
		log.Fatalln(errNewRequest)
	}

	for key, value := range eventHeaders {
		if key != "X-Sentry-Auth" {
			request.Header.Set(key, value)
		}
	}
	return request
}
