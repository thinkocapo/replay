package main

import (
	"fmt"
	"encoding/json"
	"strings"
)

func decodeEnvelope(event Event) (string, Timestamper, EnvelopeEncoder, string) {

	TRANSACTION := event.Kind == "transaction"
	JAVASCRIPT := event.Platform == "javascript"
	PYTHON := event.Platform == "python"

	storeEndpoint := matchDSN(projectDSNs, event)

	envelope := event.Body
	
	// Gotcha - Python transaction envelopes have a terminating '\n' char which causes unmarshaling to fail, "panic: unexpected end of JSON input" so remove the empty item that Splitting creates
	items := strings.Split(envelope, "\n")
	length := len(items)
	if (items[length-1] == "") {
		items = items[:length-1]
	}
	fmt.Println("\n > # of items in envelope", len(items))

	for idx, item := range items {
		fmt.Printf("\n> item %v %T \n", idx, item) // string
		// fmt.Println("> item ", item)

		var itemInterface map[string]interface{}
		if err := json.Unmarshal([]byte(item), &itemInterface); err != nil {
			panic(err)
		}
		fmt.Println("\n ITEM", item)
	}

	// I
	// GOAL
	// print/long ALL transaction types in both.json again

	// GOAL update all Timestamps and SEND

	// II
	// GOAL
	// eventId - is in first envelope item as well as largest envelope item, for both JS + PY transactions
	// 1. per item but inside 1 envelope, generate new event_id and put on both envelope items here....EASY

	// traceId - is in largest envelope item, for both JS + PY transactions
	// 1. keep a map of map[id's]itemPointersArray 2. at end, iterate through this map and give each item in itemPointersArray the same new generated Id

	// notes...
	// 1. ^ update each itemInterface in place...?
	// 2. 'OR' return envelope array-of-map[string]interfaces{} back to a string. then update
	// 3. ^ update each itemInterface in place...and put to some kind of 'output' envelope
	
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

	// var b BodyEncoder
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

// EXPERIMENT
// type decoder interface {
// 	jsError() []byte
// 	pyError() []byte
// 	envelope() []byte
// }
// func (BodyEncoder) jsError (body map[string]interface{}) []byte {
// 	return marshalJSON(body)
// }
// func (BodyEncoder) pyError (body map[string]interface{}) []byte {
// 	bodyBytes := marshalJSON(body)
// 	buf := encodeGzip(bodyBytes)
// 	return buf.Bytes()
// }
// func (BodyEncoder) envelope (envelope string) []byte {
// 	return []byte(envelope)
// }