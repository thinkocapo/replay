package main

import (
	"fmt"
	"github.com/shopspring/decimal"
	// "strconv"
	"time"
)

// https://github.com/getsentry/sentry-javascript/tree/master/packages/utils
// https://github.com/getsentry/sentry-javascript/blob/c6a2ec95f5c21df5fb6c4d7ee07087b615e23436/packages/utils/src/misc.ts
func main() {
	// price, err := decimal.NewFromString("136.02")
	// if err != nil {
	// 	panic(err)
	// }

	// quantity := decimal.NewFromInt(3)

	// fee, _ := decimal.NewFromString(".035")
	// taxRate, _ := decimal.NewFromString(".08875")

	// subtotal := price.Mul(quantity)

	// preTax := subtotal.Mul(fee.Add(decimal.NewFromFloat(1)))

	// total := preTax.Mul(taxRate.Add(decimal.NewFromFloat(1)))

	// fmt.Println("Subtotal:", subtotal)                      // Subtotal: 408.06
	// fmt.Println("Pre-tax:", preTax)                         // Pre-tax: 422.3421
	// fmt.Println("Taxes:", total.Sub(preTax))                // Taxes: 37.482861375
	// fmt.Println("Total:", total)                            // Total: 459.824961375
	// fmt.Println("Tax rate:", total.Sub(preTax).Div(preTax)) // Tax rate: 0.08875

	// try with 1591051102.7653 as well
	startTimestamp, _ := decimal.NewFromString("1591051102.765368")
	endTimestamp, _ := decimal.NewFromString("1591051102.777408")

	difference := endTimestamp.Sub(startTimestamp)
	fmt.Println("\nDifference:", difference)

	// timestamp, _ := decimal.NewFromString(string(time.Now().Unix()))
	// timestamp := decimal.NewFromFloat(float64(time.Now().Unix()))
	timestamp := time.Now().UnixNano()
	timestamp1 := fmt.Sprint(timestamp)

	// timestamp = strconv.FormatInt(timestamp, 10)
	// fmt.Println("\ntstring:", timestamp)

	fmt.Println("\ntimestamp timestamp1:", timestamp1)
	sentryTimestamp := timestamp1[:11] + "." + timestamp1[11:]
	fmt.Println("\ntimestamp         sentryTimestamp:", sentryTimestamp)

	timestamp1000 := time.Now().UnixNano()
	base := timestamp1000 / 1000000
	modulo := timestamp % 1000000
	fmt.Println("\ntimestamp1000 base:", base)
	fmt.Println("\ntimestamp1000 modulo:", modulo)

	// myv := float64(ts) // fails

	// newStartTimestamp := timestamp
	// newEndTimestamp := timestamp.Add(difference)
	// fmt.Println("\nnewStartTimestamp:", newStartTimestamp)
	// fmt.Println("\nnewEndTimestamp:", newEndTimestamp)

	// now1 := time.Now()
	// fmt.Print("time.Now().UnixNano()", now1.UnixNano())

	// nanos := now1.UnixNano()
	// logs "1591069805238407460" so could add decimal to that?
	// fmt.Println("\nnanos", nanos)
	// fmt.Println("\n", time.Unix(0, nanos))
	return
}