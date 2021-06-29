package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Request struct {
	Payload       []byte
	StoreEndpoint string
	Kind          string
	Platform      string
}

func NewRequest(event Event) *Request {
	r := new(Request)

	var bodyBytes []byte
	var err error
	if event.Kind == ERROR || event.Kind == DEFAULT {
		bodyBytes, err = json.Marshal(event.Error)
		r.Kind = ERROR
	}
	if event.Kind == TRANSACTION {
		bodyBytes, err = json.Marshal(event.Transaction)
		r.Kind = TRANSACTION
	}
	if err != nil {
		sentry.CaptureException(err)
		fmt.Println(err)
	}

	r.Platform = event.getPlatform()
	r.Payload = bodyBytes
	r.StoreEndpoint = event.DSN.storeEndpoint()

	if r.StoreEndpoint == "" || r.Payload == nil {
		sentry.CaptureException(errors.New("missing StoreEndpoint or Payload"))
		log.Fatal("missing StoreEndpoint or Payload")
	}
	return r
}

func (r Request) send() {
	time.Sleep(300 * time.Millisecond)
	var payload []byte
	size := len(r.Payload)

	HUNDRED_KILOBYTES := 100000
	if size > HUNDRED_KILOBYTES {
		fmt.Println("***** TOO BIG *****")
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		_, err := gw.Write(r.Payload)
		if err != nil {
			log.Fatal(err)
		}
		err = gw.Close()
		if err != nil {
			log.Fatal(err)
		}
		payload = buf.Bytes()
	} else {
		payload = r.Payload
	}

	request, errNewRequest := http.NewRequest("POST", r.StoreEndpoint, bytes.NewReader(payload)) // &buf
	if errNewRequest != nil {
		sentry.CaptureException(errNewRequest)
		log.Fatalln(errNewRequest)
	}

	request.Header.Set("content-type", "application/json")
	if size > HUNDRED_KILOBYTES {
		request.Header.Set("Content-Encoding", "gzip")
	}

	// fmt.Printf("\n> storeEndpoint %v\n", r.StoreEndpoint)

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
		counter++
		fmt.Printf("> Kind: %v | %v | Response: %v \n", r.Kind, r.Platform, string(responseData))
	} else {
		fmt.Printf("> event IGNORED %v | %v  \n", r.Kind, r.Platform)
	}
}
