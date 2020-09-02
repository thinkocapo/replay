package main

import (
	"fmt"
	"encoding/json"
	"strings"
)

// Encoders OG
// func envelopeEncoder(envelope string) []byte {
// 	return []byte(envelope)
// }
func envelopeEncoderJs(items []Item) []byte {
	output := []byte{}
	for _, item := range items {
		// fmt.Println("\n > envelopeEncoder", item)
		output = append(output, marshalJSONItem(item)...)
		newLine := []byte("\n")
		output = append(output, newLine...)
	}
	// fmt.Println("\n > OUTPUT ", output)
	return output
	// if err := json.Unmarshal([]byte(item), &item1); err != nil {
}
func envelopeEncoderPy(items []Item) []byte {
	output := ""
	for idx, item := range items {
		byteString, _ := json.Marshal(item)
		fmt.Println("\n > envelopeEncoder idx", idx)

		output = output + string(byteString) + "\n" // `\n` \r

		// if (len(items)-1 != idx) {
		// 	output = output + string(byteString) + "\n" // `\n` \r
		// } else {
		// 	fmt.Println("\n > FINAL")
		// 	output = output + string(byteString)
		// }
	}
	// fmt.Println("\n > envelopeEncoder OUTPUT", output)
	splitted := strings.Split(output, "\n")
	fmt.Println("\n > envelopeEncoderPy splitted length", len(splitted))

	buf := encodeGzip([]byte(output))
	return buf.Bytes()
}
// OG
// func envelopeEncoderPy(envelope string) []byte {
// 	buf := encodeGzip([]byte(envelope))
// 	return buf.Bytes()
// }


func jsEncoder(body map[string]interface{}) []byte {
	return marshalJSON(body)
}
func pyEncoder(body map[string]interface{}) []byte {
	bodyBytes := marshalJSON(body)
	buf := encodeGzip(bodyBytes)
	return buf.Bytes()
}

type BodyEncoder func(map[string]interface{}) []byte
// type EnvelopeEncoder func(string) []byte OG
type EnvelopeEncoder func(items []Item) []byte