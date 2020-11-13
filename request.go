package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Request struct {
	errorPayload       Error
	transactionPayload Transaction
	storeEndpoint      string
}

func (r Request) sendRequest() bool {

	bodyBytes, errBodyBytes := json.Marshal(r.errorPayload)
	if errBodyBytes != nil {
		fmt.Println(errBodyBytes)
	}
	request, errNewRequest := http.NewRequest("POST", r.storeEndpoint, bytes.NewReader(bodyBytes)) // &buf
	if errNewRequest != nil {
		log.Fatalln(errNewRequest)
	}
	// eventHeaders := [2]string{"content-type", "x-sentry-auth"}
	request.Header.Set("content-type", "application/json")
	fmt.Printf("*** SENTRY_AUTH_KEY ***\n", os.Getenv("SENTRY_AUTH_KEY"))
	request.Header.Set("x-sentry-auth", os.Getenv("SENTRY_AUTH_KEY"))

	// for _, key := range eventHeaders {
	// // if key != "x-Sentry-Auth" {
	// request.Header.Set(key, "asdf")
	// // }
	// }
	response, requestErr := httpClient.Do(request)
	if requestErr != nil {
		log.Fatal(requestErr)
	}
	responseData, responseDataErr := ioutil.ReadAll(response.Body)
	if responseDataErr != nil {
		log.Fatal(responseDataErr)
	}
	fmt.Printf("> KIND|RESPONSE: %s \n", string(responseData))
	return true
}

func sendRequests(requests []Request) bool {
	for _, request := range requests {
		request.sendRequest()
	}
	return true
}
