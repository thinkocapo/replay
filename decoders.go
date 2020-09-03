package main

import (
	"fmt"
	"encoding/json"
	"strings"
)

func decodeEnvelope(event Event) ([]interface{}, Timestamper, EnvelopeEncoder, string) {

	TRANSACTION := event.Kind == "transaction"
	JAVASCRIPT := event.Platform == "javascript"
	PYTHON := event.Platform == "python"

	storeEndpoint := matchDSN(projectDSNs, event)

	envelope := event.Body // fmt.Println("\n > envelope INPUT from event.Body", envelope)
	
	// Python transaction envelopes have a terminating '\n' char which causes unmarshaling to fail, "panic: unexpected end of JSON input" so remove the empty item that Splitting creates
	envelopeItems := strings.Split(envelope, "\n")
	length := len(envelopeItems)
	if (envelopeItems[length-1] == "") {
		envelopeItems = envelopeItems[:length-1]
	}

	fmt.Printf("\n > Platform %v | # of envelopeItems in envelope %v \n", event.Platform, len(envelopeItems))

	var items  []interface{}
	for idx, itemString := range envelopeItems {
		fmt.Printf("> item # %v | type %T \n", idx, itemString) // string

		var itemInterface map[string]interface{} // or interface{}?
		if err := json.Unmarshal([]byte(itemString), &itemInterface); err != nil {
			panic(err)
		}
		items = append(items, itemInterface)
	}
	
	switch {
	case JAVASCRIPT && TRANSACTION:
		return items, updateTimestamps, envelopeEncoderJs, storeEndpoint
	case PYTHON && TRANSACTION:
		return items, updateTimestamps, envelopeEncoderPy, storeEndpoint
	}

	return items, updateTimestamps, envelopeEncoderJs, storeEndpoint
}

// TODO remove 'TRANSACTION' from here
func decodeError(event Event) (map[string]interface{}, Timestamper, BodyEncoder, string) {

	body := unmarshalJSON([]byte(event.Body))

	JAVASCRIPT := event.Platform == "javascript"
	PYTHON := event.Platform == "python"
	ANDROID := event.Platform == "android"

	ERROR := event.Kind == "error"
	TRANSACTION := event.Kind == "transaction"

	storeEndpoint := matchDSN(projectDSNs, event)
	fmt.Printf("> storeEndpoint %v \n", storeEndpoint)

	// var b BodyEncoder?
	switch {
	case ANDROID && TRANSACTION:
		return body, updateTimestamp, pyEncoder, storeEndpoint
	case ANDROID && ERROR:
		return body, updateTimestamp, pyEncoder, storeEndpoint

	case JAVASCRIPT && TRANSACTION:
		return body, updateTimestamps, jsEncoder, storeEndpoint
	case JAVASCRIPT && ERROR:
		return body, updateTimestamp, jsEncoder, storeEndpoint

	case PYTHON && TRANSACTION:
		return body, updateTimestamps, pyEncoder, storeEndpoint
	case PYTHON && ERROR:
		return body, updateTimestamp, pyEncoder, storeEndpoint
	}

	return body, updateTimestamps, jsEncoder, storeEndpoint
}

type Timestamper func(map[string]interface{}, string) map[string]interface{}
