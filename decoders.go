package main

import (
	"fmt"
	"encoding/json"
	"reflect"
	"strings"
)

// func decodeEnvelope(event Event) ([]Item, Timestamper, EnvelopeEncoder, string) {
func decodeEnvelope(event Event) ([]interface{}, Timestamper, EnvelopeEncoder, string) {

	TRANSACTION := event.Kind == "transaction"
	JAVASCRIPT := event.Platform == "javascript"
	PYTHON := event.Platform == "python"

	storeEndpoint := matchDSN(projectDSNs, event)

	envelope := event.Body
	// fmt.Println("\n > envelope INPUT from event.Body", envelope)
	
	// Python transaction envelopes have a terminating '\n' char which causes unmarshaling to fail, "panic: unexpected end of JSON input" so remove the empty item that Splitting creates
	envelopeItems := strings.Split(envelope, "\n")
	length := len(envelopeItems)
	if (envelopeItems[length-1] == "") {
		envelopeItems = envelopeItems[:length-1]
	}
	fmt.Println("\n > # of envelopeItems in envelope", len(envelopeItems))

	// items := []Item{}
	var items  []interface{}

	for idx, item := range envelopeItems {
		fmt.Printf("\n> item.string %v %T \n", idx, item) // string


		// TODO if platform==python then treat it one way, if platform==javascript then treat it another
		// Read through item string....
			// if ever 9 numbers in a row
			// then it's Item2{}

		// item1 := Item{}
		// item2 := Item2{}

		// if err := json.Unmarshal([]byte(item), &item1); err != nil {
		// 	fmt.Println("\n > There was an error but will try something else...")
		// 	// panic(err)

		// 	if err2 := json.Unmarshal([]byte(item), &item2); err2 != nil {
		// 		fmt.Println("\n > Here we are ")
		// 		panic(err2)
		// 	}
		// }

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(item), &parsed); err != nil {
			panic(err)
		}
		// fmt.Println("parsed.platform", parsed["platform"])

		// fmt.Println("PARSED", parsed)

		if val, ok := parsed["timestamp"]; ok {
			fmt.Println("\n > parsed[timestamp]", val)

			switch reflect.TypeOf(parsed["timestamp"]).String() {
			case "float64":
				fmt.Println("\n > THAT WAS PYTHON")
			case "string":
				fmt.Println("\n > THAT WAS JAVASCRIPT")
			default:
				panic("JSON type is not understood")
			}
		} else {
			// parse as regular Item{}
			fmt.Println("\n > no timestamp, must be a header")
		}

		// TODO return interface{}, and do Item{} type later, or never!
		items = append(items, parsed)
		// items = append(items, item1)
	}
	fmt.Println("\n # of items in []Item{}", len(items))

	// update all Timestamps and SEND

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