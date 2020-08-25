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

func unmarshalEnvelope(bytes []byte) []string {
	var envelope string
	envelope = string(bytes)
	envelopeContents := strings.Split(envelope, "\n")

	fmt.Print("\n . . . . . . . . UNMARSHAL ENVELOPE . . . . . . . . . . . ")
	// fmt.Print(envelopeContents[0], "\n")

	// TODO make into object map[string]interface{}


	// fmt.Print(envelopeContents[1], "\n")
	// fmt.Print(envelopeContents[2], "\n")

	// if err := json.Unmarshal(bytes, &text); err != nil {
	// 	panic(err)
	// }
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
