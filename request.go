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
	EventJson
	transactionPayload Transaction
	storeEndpoint      string
}

func (r Request) sendRequest(ignore bool) bool {

	// bodyBytes, errBodyBytes := json.Marshal(r.errorPayload)
	bodyBytes, errBodyBytes := json.Marshal(r.Error)
	if errBodyBytes != nil {
		fmt.Println(errBodyBytes)
	}
	request, errNewRequest := http.NewRequest("POST", r.storeEndpoint, bytes.NewReader(bodyBytes)) // &buf
	if errNewRequest != nil {
		log.Fatalln(errNewRequest)
	}

	// TODO 12;33p if both appearing, then don't need this:
	// request.Header.Set("x-sentry-auth", os.Getenv("SENTRY_AUTH_KEY"))
	// fmt.Printf("\n> x-sentry-auth %v\n", os.Getenv("SENTRY_AUTH_KEY"))

	request.Header.Set("content-type", "application/json")
	fmt.Printf("\n> storeEndpoint %v\n", r.storeEndpoint)

	// TODO remove this, b/c using UnmarshalJSON
	// fmt.Printf("\n> errorPayload %+v \n", r.errorPayload)
	fmt.Printf("\n> errorPayload %+v \n", r.Error)

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

// TODO rename file to requests.send() ?
// has access to event and storeEndpoint, so could do call dsnToStoreEndpoint() from here...
// iterates through `for event := range requests.events`
// request = Request{
// 	EventJson:     event,
// 	storeEndpoint: dsnToStoreEndpoint(projectDSNs, event.Error.Platform),
// })
//request.send()
