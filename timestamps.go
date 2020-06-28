package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
)

func updateTimestamp(bodyInterface map[string]interface{}, platform string) map[string]interface{} {
	fmt.Println("> Error timestamp before", bodyInterface["timestamp"])
	bodyInterface["timestamp"] = time.Now().Unix()
	fmt.Println("> Error timestamp after ", bodyInterface["timestamp"])

	fmt.Println("platform string", platform)
	return bodyInterface
}

// used for TRANSACTIONS
// start/end here is same as the sdk's start_timestamp/timestamp, and start_timestamp is only present in transactions
// For future reference, data.contexts.trace.span_id is the Parent Trace and at one point I thoguht I saw data.entries with spans. Disregarding it for now.
// Subtraction arithmetic needed on the decimals via Floats, so avoid Int's
// Better to put as Float64 before serialization. also keep to 7 decimal places as the range sent by sdk's is 4 to 7
func updateTimestamps(data map[string]interface{}, platform string) map[string]interface{} {
	fmt.Printf("\n> both updateTimestamps PARENT start_timestamp before %v (%T) \n", data["start_timestamp"], data["start_timestamp"])
	fmt.Printf("> both updateTimestamps PARENT       timestamp before %v (%T)", data["timestamp"], data["timestamp"])

	var parentStartTimestamp, parentEndTimestamp decimal.Decimal
	// PYTHON timestamp format is 2020-06-06T04:54:56.636664Z RFC3339Nano
	if platform == "python" {
		parentStart, _ := time.Parse(time.RFC3339Nano, data["start_timestamp"].(string)) // integer?
		parentEnd, _ := time.Parse(time.RFC3339Nano, data["timestamp"].(string))
		parentStartTime := fmt.Sprint(parentStart.UnixNano())
		parentEndTime := fmt.Sprint(parentEnd.UnixNano())
		parentStartTimestamp, _ = decimal.NewFromString(parentStartTime[:10] + "." + parentStartTime[10:])
		parentEndTimestamp, _ = decimal.NewFromString(parentEndTime[:10] + "." + parentEndTime[10:])
	}
	// JAVASCRIPT timestamp format is 1591419091.4805 to 1591419092.000035
	if platform == "javascript" {
		// in sqlite it was float64, not a string. or rather, Go is making it a float64 upon reading from db? not sure
		// make into a 'decimal' class type for logging or else it logs as "1.5914674155654302e+09" instead of 1591467415.5654302
		parentStartTimestamp = decimal.NewFromFloat(data["start_timestamp"].(float64))
		parentEndTimestamp = decimal.NewFromFloat(data["timestamp"].(float64))
	}

	// PARENT TRACE
	// Adjust the parentDifference/spanDifference between .01 and .2 (1% and 20% difference) so the 'end timestamp's always shift the same amount (no gaps at the end)
	parentDifference := parentEndTimestamp.Sub(parentStartTimestamp)
	fmt.Print("\n> parentDifference before", parentDifference)
	rand.Seed(time.Now().UnixNano())
	percentage := 0.01 + rand.Float64()*(0.20-0.01)
	fmt.Println("\n> percentage", percentage)
	rate := decimal.NewFromFloat(percentage)
	parentDifference = parentDifference.Mul(rate.Add(decimal.NewFromFloat(1)))
	fmt.Print("\n> parentDifference after", parentDifference)

	unixTimestampString := fmt.Sprint(time.Now().UnixNano())
	newParentStartTimestamp, _ := decimal.NewFromString(unixTimestampString[:10] + "." + unixTimestampString[10:])
	newParentEndTimestamp := newParentStartTimestamp.Add(parentDifference)

	if !newParentEndTimestamp.Sub(newParentStartTimestamp).Equal(parentDifference) {
		fmt.Print("\nFALSE - parent BOTH", newParentEndTimestamp.Sub(newParentStartTimestamp))
	}

	data["start_timestamp"], _ = newParentStartTimestamp.Round(7).Float64()
	data["timestamp"], _ = newParentEndTimestamp.Round(7).Float64()

	// Could conver back to RFC3339Nano (as that's what the python sdk uses for transactions Python Transactions use) but Floats are working and mirrors what the javascript() function does
	// logging with decimal just so it's more readable and convertible in https://www.epochconverter.com/, because the 'Float' form is like 1.5914674155654302e+09
	fmt.Printf("\n> both updateTimestamps PARENT start_timestamp after %v (%T) \n", decimal.NewFromFloat(data["start_timestamp"].(float64)), data["start_timestamp"])
	fmt.Printf("> both updateTimestamps PARENT       timestamp after %v (%T) \n", decimal.NewFromFloat(data["timestamp"].(float64)), data["timestamp"])

	// SPAN
	// TEST for making sure that the span object was updated by reference
	// firstSpan := data["spans"].([]interface{})[0].(map[string]interface{})
	// fmt.Printf("\n> before ", decimal.NewFromFloat(firstSpan["start_timestamp"].(float64)))
	for _, span := range data["spans"].([]interface{}) {
		sp := span.(map[string]interface{})
		// fmt.Printf("\n> both updatetimestamps SPAN start_timestamp before %v (%T)", sp["start_timestamp"], sp["start_timestamp"])
		// fmt.Printf("\n> both updatetimestamps SPAN       timestamp before %v (%T)\n", sp["timestamp"]	, sp["timestamp"])

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
		fmt.Println("> spanDifference before", spanDifference)
		spanDifference = spanDifference.Mul(rate.Add(decimal.NewFromFloat(1)))
		fmt.Println("> spanDifference after", spanDifference)

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

		// logging with decimal just so it's more readable and convertible in https://www.epochconverter.com/, because the 'Float' form is like 1.5914674155654302e+09
		fmt.Printf("\n> both updatetimestamps SPAN start_timestamp after %v (%T)", decimal.NewFromFloat(sp["start_timestamp"].(float64)), sp["start_timestamp"])
		fmt.Printf("\n> both updatetimestamps SPAN       timestamp after %v (%T)\n", decimal.NewFromFloat(sp["timestamp"].(float64)), sp["timestamp"])
	}
	// TESt for making sure that the span object was updated by reference. E.g. 1591467416.0387652 should now be 1591476953.491206959
	// fmt.Printf("\n> after ", firstSpan["start_timestamp"])
	return data
}
