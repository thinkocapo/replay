package main

import (
	"fmt"
	"strings"
)

func decodeEnvelope(event Event) (string, Timestamper, EnvelopeEncoder, string) {

	TRANSACTION := event.Kind == "transaction"
	JAVASCRIPT := event.Platform == "javascript"
	PYTHON := event.Platform == "python"

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
		return envelope, updateTimestamps, envelopeEncoder, storeEndpoint
	case PYTHON && TRANSACTION:
		return envelope, updateTimestamps, envelopeEncoderPy, storeEndpoint
	}

	return envelope, updateTimestamps, envelopeEncoder, storeEndpoint
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

// Encoders
func envelopeEncoder(envelope string) []byte {
	return []byte(envelope)
}
func envelopeEncoderPy(envelope string) []byte {
	buf := encodeGzip([]byte(envelope))
	return buf.Bytes()
}
func jsEncoder(body map[string]interface{}) []byte {
	return marshalJSON(body)
}
func pyEncoder(body map[string]interface{}) []byte {
	bodyBytes := marshalJSON(body)
	buf := encodeGzip(bodyBytes)
	return buf.Bytes()
}

type BodyEncoder func(map[string]interface{}) []byte
type EnvelopeEncoder func(string) []byte
type Timestamper func(map[string]interface{}, string) map[string]interface{}