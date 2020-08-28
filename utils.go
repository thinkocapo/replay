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
func decodeEnvelope(event Event) (string, Timestamper, BodyEncoder, []string, string) {

	TRANSACTION := event.Kind == "transaction"
	JAVASCRIPT := event.Platform == "javascript"
	PYTHON := event.Platform == "python"
	jsHeaders := []string{"Accept-Encoding", "Content-Length", "Content-Type", "User-Agent"}
	pyHeaders := []string{"Accept-Encoding", "Content-Length", "Content-Encoding", "Content-Type", "User-Agent"}
	storeEndpoint := matchDSN(projectDSNs, event)
	
	fmt.Printf("> storeEndpoint1 %v \n", storeEndpoint)

	envelope := event.Body
	items := strings.Split(envelope, "\n")
	var item map[string]interface{}

	fmt.Println("\n > # of items in envelope", len(items))
	for idx, _ := range items {
		fmt.Println("\n > item is...", idx)
	}
	
	// TODO need do this for every item in items
	if err := json.Unmarshal([]byte(items[0]), &item); err != nil {
		panic(err)
	}

	switch {
	case JAVASCRIPT && TRANSACTION:
		return envelope, updateTimestamps, jsEncoder, jsHeaders, storeEndpoint
	case PYTHON && TRANSACTION:
		return envelope, updateTimestamps, jsEncoder, pyHeaders, storeEndpoint // because envelope so jsEncoder....?
	}

	// TODO return array of map[string]interface{}? where to update envelope items? timestamps, ID's
	return envelope, updateTimestamps, jsEncoder, jsHeaders, storeEndpoint
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
