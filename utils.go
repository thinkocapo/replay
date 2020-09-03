package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func marshalJSON(body map[string]interface{}) []byte {
	bodyBytes, errBodyBytes := json.Marshal(body)
	if errBodyBytes != nil {
		fmt.Println(errBodyBytes)
	}
	return bodyBytes
}

func marshalJSONItem(item interface{}) []byte {
	//fmt.Println("\n > marshalJSONItem BEFORE", item) // {2b7e81ebe33349cda2a77f04c30e8174 2020-08-29T05:43:31.286573Z}
	itemBytes, errItemBytes := json.Marshal(item)
	if errItemBytes != nil {
		fmt.Println(errItemBytes)
	}
	//fmt.Println("\n > marshalJSONItem ", string(itemBytes)) // {"event_id":"2b7e81ebe33349cda2a77f04c30e8174","sent_at":"2020-08-29T05:43:31.286573Z"}
	return itemBytes
}

func unmarshalJSON(bytes []byte) map[string]interface{} {
	var _interface map[string]interface{}
	if err := json.Unmarshal(bytes, &_interface); err != nil {
		panic(err)
	}
	return _interface
}

// func unmarshalJSONItem(item string) Item {
// 	var _interface Item
// 	if err := json.Unmarshal([]byte(item), &_interface); err != nil {
// 		panic(err)
// 	}
// 	return _interface
// 	// return Item{id: "test"} // wouldn't work
// }

