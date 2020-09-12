package undertaker

import (
	"encoding/json"
)

func envelopeEncoderJs(items []interface{}) []byte {
	output := []byte{}
	for _, item := range items {
		output = append(output, marshalJSONItem(item)...)
		newLine := []byte("\n")
		output = append(output, newLine...)
	}
	return output
}

func envelopeEncoderPy(items []interface{}) []byte {
	output := ""
	for _, item := range items {
		byteString, _ := json.Marshal(item)
		output = output + string(byteString) + "\n" // `\n` \r
		// if (len(items)-1 != idx) {
		// 	output = output + string(byteString) + "\n" // `\n` \r
		// } else {
		// 	fmt.Println("\n > FINAL")
		// 	output = output + string(byteString)
		// }
	}
	// splitted := strings.Split(output, "\n")
	// fmt.Println("\n > envelopeEncoderPy splitted length", len(splitted))

	buf := encodeGzip([]byte(output))
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
type EnvelopeEncoder func(items []interface{}) []byte
