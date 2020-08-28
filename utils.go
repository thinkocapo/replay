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
func decodeEnvelope(envelope string) []string {

	items := strings.Split(envelope, "\n")

	var item map[string]interface{}

	fmt.Println("\n > # of items in envelope", len(items))

	for idx, item := range items {
		fmt.Println("\n > item is...", idx)
		fmt.Println(item)
	}
	
	// shouldn't need, since fixing encoding problem
	// stripped := strings.ReplaceAll(items[0], "\\", "")
	// remove the prepending quotation mark on "{\"event_id\": so it becomes {\"event_id\"
	// stripped = stripped[1:]
	fmt.Println("\n0000000 . . ")
	// TODO need do this for every item in items
	if err := json.Unmarshal([]byte(items[0]), &item); err != nil {
		fmt.Println("111111. . . ")
		panic(err)
	}
	fmt.Println("2222. . . .")

	fmt.Println("\n > ITEM example", item)

	// TODO return array of map[string]interface{}
	return items
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
