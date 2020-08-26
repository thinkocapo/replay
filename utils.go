package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

func decodeGzip(bodyBytesInput []byte) (bodyBytesOutput []byte) {
	bodyReader, err := gzip.NewReader(bytes.NewReader(bodyBytesInput))
	if err != nil {
		fmt.Println(err)
	}
	bodyBytesOutput, err = ioutil.ReadAll(bodyReader)
	if err != nil {
		fmt.Println(err)
	}
	return
}

func encodeGzip(b []byte) bytes.Buffer {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(b)
	w.Close()
	// return buf.Bytes()
	return buf
}

// func unmarshalEnvelope(bytes []byte) []string {
func unmarshalEnvelope(envelope string) []string {

	// fmt.Println("\n . . . . . . . . UNMARSHAL ENVELOPE . . . . . . . . . . .", len(envelopeContents))
	// var envelope string
	// envelope = string(bytes)
	envelopeContents := strings.Split(envelope, "\\n")

	var content map[string]interface{}

	fmt.Println(envelopeContents[0]) 
	
	stripped := strings.ReplaceAll(envelopeContents[0], "\\", "")

	// remove the prepending quotation mark on "{\"event_id\": so it becomes {\"event_id\"
	stripped = stripped[1:]

	if err := json.Unmarshal([]byte(stripped), &content); err != nil {
		panic(err)
	}
	// SUCCESS
	fmt.Println("CONTENT", content)
	
	return envelopeContents
}

func unmarshalJSON(bytes []byte) map[string]interface{} {
	var _interface map[string]interface{}
	if err := json.Unmarshal(bytes, &_interface); err != nil {
		panic(err)
	}
	return _interface
}

func marshalJSON(body map[string]interface{}) []byte {
	bodyBytes, errBodyBytes := json.Marshal(body)
	if errBodyBytes != nil {
		fmt.Println(errBodyBytes)
	}
	return bodyBytes
}
