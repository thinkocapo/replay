package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Request struct {
	errorPayload       Error
	transactionPayload Transaction
	storeEndpoint      string
}

func (r Request) sendRequest(ignore bool) bool {

	bodyBytes, errBodyBytes := json.Marshal(r.errorPayload)
	if errBodyBytes != nil {
		fmt.Println(errBodyBytes)
	}
	request, errNewRequest := http.NewRequest("POST", r.storeEndpoint, bytes.NewReader(bodyBytes)) // &buf
	if errNewRequest != nil {
		log.Fatalln(errNewRequest)
	}

	request.Header.Set("x-sentry-auth", os.Getenv("SENTRY_AUTH_KEY"))
	request.Header.Set("content-type", "application/json")

	fmt.Printf("\n> x-sentry-auth %v\n", os.Getenv("SENTRY_AUTH_KEY"))
	fmt.Printf("\n> storeEndpoint %v\n", r.storeEndpoint)
	fmt.Printf("\n> errorPayload %+v \n", r.errorPayload)

	if !ignore {
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

func sendRequests(requests []Request, ignore bool) bool {
	for _, request := range requests {
		// if !ignore {
		request.sendRequest(ignore)
		// } else {
		// fmt.Printf("> %s event IGNORED \n", request.storeEndpoint)
		// }
		time.Sleep(750 * time.Millisecond)
	}
	return true
}
