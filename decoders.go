package main

import (
	// "bytes"
	// "compress/gzip"
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	"strings"
)

func decodeEnvelope(event Event) (string, Timestamper, EnvelopeEncoder, []string, string) {

	TRANSACTION := event.Kind == "transaction"
	JAVASCRIPT := event.Platform == "javascript"
	PYTHON := event.Platform == "python"
	jsHeaders := []string{"Accept-Encoding", "Content-Length", "Content-Type", "User-Agent"}
	pyHeaders := []string{"Accept-Encoding", "Content-Length", "Content-Encoding", "Content-Type", "User-Agent"}
	storeEndpoint := matchDSN(projectDSNs, event)

	fmt.Printf("> storeEndpoint %v \n", storeEndpoint)

	envelope := event.Body

	items := strings.Split(envelope, "\n")
	fmt.Println("\n > # of items in envelope", len(items))
	for idx, _ := range items {
		fmt.Println("> item", idx)
		// TODO need do this for every item in items
		// var item map[string]interface{}
		// if err := json.Unmarshal([]byte(items[0]), &item); err != nil {
			// panic(err)
		// }
	}

	// TODO return envelope array-of-map[string]interfaces{} back to a string
	// TODO return bodyEncoder for []byte(envelope) maybe called 'envelopeEncoder'. Go strings are already utf-8 encoded
	
	switch {
	case JAVASCRIPT && TRANSACTION:
		return envelope, updateTimestamps, envelopeEncoder, jsHeaders, storeEndpoint
	case PYTHON && TRANSACTION:
		return envelope, updateTimestamps, envelopeEncoderPy, pyHeaders, storeEndpoint // because envelope so jsEncoder....?
	}

	return envelope, updateTimestamps, envelopeEncoder, jsHeaders, storeEndpoint
}

// TODO remove 'TRANSACTION' from here
func decodeEvent(event Event) (map[string]interface{}, Timestamper, BodyEncoder, []string, string) {

	body := unmarshalJSON([]byte(event.Body))

	JAVASCRIPT := event.Platform == "javascript"
	PYTHON := event.Platform == "python"
	ANDROID := event.Platform == "android"

	ERROR := event.Kind == "error"
	TRANSACTION := event.Kind == "transaction"

	jsHeaders := []string{"Accept-Encoding", "Content-Length", "Content-Type", "User-Agent"}
	pyHeaders := []string{"Accept-Encoding", "Content-Length", "Content-Encoding", "Content-Type", "User-Agent"}
	androidHeaders := []string{"Content-Length","User-Agent","Connection","Content-Encoding","X-Forwarded-Proto","Host","Accept","X-Forwarded-For"} // X-Sentry-Auth omitted
	storeEndpoint := matchDSN(projectDSNs, event)
	fmt.Printf("> storeEndpoint %v \n", storeEndpoint)

	switch {
	case ANDROID && TRANSACTION:
		return body, updateTimestamp, pyEncoder, androidHeaders, storeEndpoint
	case ANDROID && ERROR:
		return body, updateTimestamp, pyEncoder, androidHeaders, storeEndpoint

	case JAVASCRIPT && TRANSACTION:
		return body, updateTimestamps, jsEncoder, jsHeaders, storeEndpoint
	case JAVASCRIPT && ERROR:
		return body, updateTimestamp, jsEncoder, jsHeaders, storeEndpoint

	case PYTHON && TRANSACTION:
		return body, updateTimestamps, pyEncoder, pyHeaders, storeEndpoint
	case PYTHON && ERROR:
		return body, updateTimestamp, pyEncoder, pyHeaders, storeEndpoint
	}

	return body, updateTimestamps, jsEncoder, jsHeaders, storeEndpoint
}