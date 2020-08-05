package undertaker

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
)

/*
PYTHON timestamp format is 2020-06-06T04:54:56.636664Z RFC3339Nano
JAVASCRIPT timestamp format is 1591419091.4805 to 1591419092.000035
PARENT TRACE - Adjust the parentDifference/spanDifference between .01 and .2 (1% and 20% difference) so the 'end timestamp's always shift the same amount (no gaps at the end)
TRANSACTIONS. body.contexts.trace.span_id is the Parent Trace. start/end here is same as the sdk's start_timestamp/timestamp, and start_timestamp is only present in transactions
To see a full span `firstSpan := body["spans"].([]interface{})[0].(map[string]interface{})``
7 decimal places as the range sent by sdk's is 4 to 7
https://www.epochconverter.com/
Float form is like 1.5914674155654302e+09
*/

// Errors
func updateTimestamp(body map[string]interface{}, platform string) map[string]interface{} {
	// fmt.Println("> Error timestamp before", body["timestamp"])
	body["timestamp"] = time.Now().Unix()
	// fmt.Println("> Error timestamp after ", body["timestamp"])
	return body
}

// Transactions - keep start and end timestamps relative to each other by computing the difference and new timestamps based on that
func updateTimestamps(body map[string]interface{}, platform string) map[string]interface{} {
	// fmt.Printf("\n> updateTimestamps PARENT start_timestamp before %v (%T) \n", body["start_timestamp"], body["start_timestamp"])
	// fmt.Printf("> updateTimestamps PARENT       timestamp before %v (%T)", body["timestamp"], body["timestamp"])

	var parentStartTimestamp, parentEndTimestamp decimal.Decimal
	if platform == "python" {
		parentStart, _ := time.Parse(time.RFC3339Nano, body["start_timestamp"].(string)) // integer?
		parentEnd, _ := time.Parse(time.RFC3339Nano, body["timestamp"].(string))
		parentStartTime := fmt.Sprint(parentStart.UnixNano())
		parentEndTime := fmt.Sprint(parentEnd.UnixNano())
		parentStartTimestamp, _ = decimal.NewFromString(parentStartTime[:10] + "." + parentStartTime[10:])
		parentEndTimestamp, _ = decimal.NewFromString(parentEndTime[:10] + "." + parentEndTime[10:])
	}
	if platform == "javascript" {
		parentStartTimestamp = decimal.NewFromFloat(body["start_timestamp"].(float64))
		parentEndTimestamp = decimal.NewFromFloat(body["timestamp"].(float64))
	}

	// Parent Trace
	parentDifference := parentEndTimestamp.Sub(parentStartTimestamp)
	// fmt.Print("\n> parentDifference before", parentDifference)
	rand.Seed(time.Now().UnixNano())
	percentage := 0.01 + rand.Float64()*(0.20-0.01)
	// fmt.Println("\n> percentage", percentage)
	rate := decimal.NewFromFloat(percentage)
	parentDifference = parentDifference.Mul(rate.Add(decimal.NewFromFloat(1)))
	// fmt.Print("\n> parentDifference after", parentDifference)

	unixTimestampString := fmt.Sprint(time.Now().UnixNano())
	newParentStartTimestamp, _ := decimal.NewFromString(unixTimestampString[:10] + "." + unixTimestampString[10:])
	newParentEndTimestamp := newParentStartTimestamp.Add(parentDifference)

	if !newParentEndTimestamp.Sub(newParentStartTimestamp).Equal(parentDifference) {
		fmt.Print("\nFALSE - parent BOTH", newParentEndTimestamp.Sub(newParentStartTimestamp))
	}

	body["start_timestamp"], _ = newParentStartTimestamp.Round(7).Float64()
	body["timestamp"], _ = newParentEndTimestamp.Round(7).Float64()

	// Could conver back to RFC3339Nano (as that's what the python sdk uses for transactions Python Transactions use) but Floats are working and this is what happens in Javascript
	// logging with 'decimal type for readability and convertability
	// fmt.Printf("\n> updateTimestamps PARENT start_timestamp after %v (%T) \n", decimal.NewFromFloat(body["start_timestamp"].(float64)), body["start_timestamp"])
	// fmt.Printf("> updateTimestamps PARENT       timestamp after %v (%T) \n", decimal.NewFromFloat(body["timestamp"].(float64)), body["timestamp"])

	// Span(s)
	for _, span := range body["spans"].([]interface{}) {
		sp := span.(map[string]interface{})
		// fmt.Printf("\n> updatetimestamps SPAN start_timestamp before %v (%T)", sp["start_timestamp"], sp["start_timestamp"])
		// fmt.Printf("\n> updatetimestamps SPAN       timestamp before %v (%T)\n", sp["timestamp"]	, sp["timestamp"])
		var spanStartTimestamp, spanEndTimestamp decimal.Decimal
		if platform == "python" {
			spanStart, _ := time.Parse(time.RFC3339Nano, sp["start_timestamp"].(string))
			spanEnd, _ := time.Parse(time.RFC3339Nano, sp["timestamp"].(string))
			spanStartTime := fmt.Sprint(spanStart.UnixNano())
			spanEndTime := fmt.Sprint(spanEnd.UnixNano())
			spanStartTimestamp, _ = decimal.NewFromString(spanStartTime[:10] + "." + spanStartTime[10:])
			spanEndTimestamp, _ = decimal.NewFromString(spanEndTime[:10] + "." + spanEndTime[10:])
		}
		if platform == "javascript" {
			spanStartTimestamp = decimal.NewFromFloat(sp["start_timestamp"].(float64))
			spanEndTimestamp = decimal.NewFromFloat(sp["timestamp"].(float64))
		}

		spanDifference := spanEndTimestamp.Sub(spanStartTimestamp)
		// fmt.Println("> spanDifference before", spanDifference)
		spanDifference = spanDifference.Mul(rate.Add(decimal.NewFromFloat(1)))
		// fmt.Println("> spanDifference after", spanDifference)

		spanToParentDifference := spanStartTimestamp.Sub(parentStartTimestamp)

		// should use newParentStartTimestamp instead of spanStartTimestamp?
		unixTimestampString := fmt.Sprint(time.Now().UnixNano())
		unixTimestampDecimal, _ := decimal.NewFromString(unixTimestampString[:10] + "." + unixTimestampString[10:])
		newSpanStartTimestamp := unixTimestampDecimal.Add(spanToParentDifference)
		newSpanEndTimestamp := newSpanStartTimestamp.Add(spanDifference)

		if !newSpanEndTimestamp.Sub(newSpanStartTimestamp).Equal(spanDifference) {
			fmt.Print("\nFALSE - span BOTH", newSpanEndTimestamp.Sub(newSpanStartTimestamp))
		}

		sp["start_timestamp"], _ = newSpanStartTimestamp.Round(7).Float64()
		sp["timestamp"], _ = newSpanEndTimestamp.Round(7).Float64()

		// logging with decimal for readability and convertability
		// fmt.Printf("\n> updatetimestamps SPAN start_timestamp after %v (%T)", decimal.NewFromFloat(sp["start_timestamp"].(float64)), sp["start_timestamp"])
		// fmt.Printf("\n> updatetimestamps SPAN       timestamp after %v (%T)\n", decimal.NewFromFloat(sp["timestamp"].(float64)), sp["timestamp"])
	}
	return body
}
