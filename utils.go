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
	envelopeContents := strings.Split(envelope, "\\n")
	fmt.Println("\n . . . . . . . . UNMARSHAL ENVELOPE . . . . . . . . . . .", len(envelopeContents))

	var content map[string]interface{}
	// var content map[string]string

	fmt.Println(envelopeContents[0]) // [1:] if extra "" at the beginning
	
	ans := strings.ReplaceAll(envelopeContents[0], "\\", "")
	// ans = ans + "\""
	ans = ans[1:]
	fmt.Println("answer", ans)

	// if err := json.Unmarshal([]byte(envelopeContents[0][1:]), &content); err != nil {
	if err := json.Unmarshal([]byte(ans), &content); err != nil {
		panic(err)
	}
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
