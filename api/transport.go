package undertaker

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// TODO maybe?
type Theencoder interface {
	encodeit() []byte
}

// TODO factory? how to rename?
type Transport struct {
	kind            string
	platform        string
	eventHeaders    map[string]string
	storeEndpoint   string
	encoded         []byte
	bodyError       map[string]interface{}
	bodyEncoder     BodyEncoder
	envelopeItems   []interface{}
	envelopeEncoder EnvelopeEncoder
}

func encodeAndSendEvents(requests []Transport) {

	for _, transport := range requests {
		if transport.kind == "transaction" {
			transport.encoded = transport.envelopeEncoder(transport.envelopeItems)
		}
		if transport.kind == "error" {
			transport.encoded = transport.bodyEncoder(transport.bodyError)
		}
		request := buildRequest(transport.encoded, transport.eventHeaders, transport.storeEndpoint)

		response, requestErr := httpClient.Do(request)
		if requestErr != nil {
			log.Fatal(requestErr)
		}
		responseData, responseDataErr := ioutil.ReadAll(response.Body)
		if responseDataErr != nil {
			log.Fatal(responseDataErr)
		}
		fmt.Printf("> KIND|RESPONSE: %s %s\n", transport.kind, string(responseData))

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
