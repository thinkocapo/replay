package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// 'sent_at' and 'length' are not on the post-ingested event. They were on the pre-ingested and had to be updated.
type Transaction struct {
	EventId   string                 `json:"event_id"`
	Release   string                 `json:"release"`
	User      map[string]interface{} `json:"user"`
	Timestamp float64                `json:"timestamp"`
	Type      string                 `json:"type"`
	Platform  string                 `json:"platform"`

	Project         int                      `json:"project"`
	Message         string                   `json:"message"`
	Tags            [][]string               `json:"tags"`
	Breadcrumbs     map[string]interface{}   `json:"breadcrumbs"`
	Contexts        map[string]interface{}   `json:"contexts"`
	Culprit         string                   `json:"culprit"`
	Environment     string                   `json:"environment"`
	Extra           map[string]interface{}   `json:"extra"`
	Grouping_config map[string]interface{}   `json:"grouping_config"`
	Key_id          string                   `json:"key_id"`
	Level           string                   `json:"level"`
	Logger          string                   `json:"logger"`
	Metadata        map[string]interface{}   `json:"metadata"`
	Received        float64                  `json:"received"`
	Request         map[string]interface{}   `json:"request"`
	Sdk             map[string]interface{}   `json:"sdk"`
	Version         string                   `json:"version"`
	Spans           []map[string]interface{} `json:"spans"`
	Start_timestamp float64                  `json:"start_timestamp"`
	Transaction     string                   `json:"transaction"`
}

const TRANSACTION = "transaction"

func (t *Transaction) eventId() {
	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
	t.EventId = uuid4
}

func (t *Transaction) release() {
	date := time.Now()
	month := date.Month()
	day := date.Day()
	var week int
	switch {
	case day <= 7:
		week = 1
	case day >= 8 && day <= 14:
		week = 2
	case day >= 15 && day <= 21:
		week = 3
	case day >= 22:
		week = 4
	}
	release := fmt.Sprint(int(month), ".", week)
	t.Release = release
}

func (t *Transaction) user() {
	t.User = make(map[string]interface{})
	user := t.User
	user["email"] = createUser()
}

func (t *Transaction) timestamps() {
	// fmt.Printf("\n> updateTimestamps PARENT start_timestamp before %v (%T) ", t.Start_timestamp, t.Start_timestamp)
	// fmt.Printf("\n> updateTimestamps PARENT       timestamp before %v (%T) \n", t.Timestamp, t.Timestamp)

	if t.Timestamp != 0 && t.Start_timestamp != 0 {
		var parentStartTimestamp, parentEndTimestamp decimal.Decimal
		if t.Platform == "python" {
			parentStartTimestamp = decimal.NewFromFloat(t.Start_timestamp)
			parentEndTimestamp = decimal.NewFromFloat(t.Timestamp)
		}
		if t.Platform == "javascript" {
			parentStartTimestamp = decimal.NewFromFloat(t.Start_timestamp)
			parentEndTimestamp = decimal.NewFromFloat(t.Timestamp)
		}

		// TRACE PARENT
		parentDifference := parentEndTimestamp.Sub(parentStartTimestamp)
		rand.Seed(time.Now().UnixNano())
		percentage := 0.01 + rand.Float64()*(0.20-0.01)
		rate := decimal.NewFromFloat(percentage)
		parentDifference = parentDifference.Mul(rate.Add(decimal.NewFromFloat(1)))

		unixTimestampString := fmt.Sprint(time.Now().UnixNano())
		newParentStartTimestamp, _ := decimal.NewFromString(unixTimestampString[:10] + "." + unixTimestampString[10:])

		newParentEndTimestamp := newParentStartTimestamp.Add(parentDifference)

		if !newParentEndTimestamp.Sub(newParentStartTimestamp).Equal(parentDifference) {
			fmt.Print("\nFALSE - parent BOTH", newParentEndTimestamp.Sub(newParentStartTimestamp))
		}

		t.Start_timestamp, _ = newParentStartTimestamp.Round(7).Float64()
		t.Timestamp, _ = newParentEndTimestamp.Round(7).Float64()

		// fmt.Printf("> updateTimestamps PARENT start_timestamp after %v (%T) \n", decimal.NewFromFloat(t.Start_timestamp), t.Start_timestamp)
		// fmt.Printf("> updateTimestamps PARENT       timestamp after %v (%T) \n", decimal.NewFromFloat(t.Timestamp), t.Timestamp)

		// SPANS
		for _, span := range t.Spans {
			var spanStartTimestamp, spanEndTimestamp decimal.Decimal
			// fmt.Printf("\n> updatetimestamps SPAN start_timestamp before %v (%T)", span["start_timestamp"].(float64), span["start_timestamp"])
			// fmt.Printf("\n> updatetimestamps SPAN       timestamp before %v (%T)\n", span["timestamp"].(float64), span["timestamp"])

			if t.Platform == "python" {
				spanStartTimestamp = decimal.NewFromFloat(span["start_timestamp"].(float64))
				spanEndTimestamp = decimal.NewFromFloat(span["timestamp"].(float64))
			}
			if t.Platform == "javascript" {
				spanStartTimestamp = decimal.NewFromFloat(span["start_timestamp"].(float64))
				spanEndTimestamp = decimal.NewFromFloat(span["timestamp"].(float64))
			}
			spanDifference := spanEndTimestamp.Sub(spanStartTimestamp)
			spanDifference = spanDifference.Mul(rate.Add(decimal.NewFromFloat(1)))

			spanToParentDifference := spanStartTimestamp.Sub(parentStartTimestamp)
			spanToParentDifference = spanToParentDifference.Mul(rate.Add(decimal.NewFromFloat(1)))

			unixTimestampString := fmt.Sprint(time.Now().UnixNano())
			unixTimestampDecimal, _ := decimal.NewFromString(unixTimestampString[:10] + "." + unixTimestampString[10:])
			newSpanStartTimestamp := unixTimestampDecimal.Add(spanToParentDifference)
			newSpanEndTimestamp := newSpanStartTimestamp.Add(spanDifference)

			if !newSpanEndTimestamp.Sub(newSpanStartTimestamp).Equal(spanDifference) {
				fmt.Print("\nFALSE - span BOTH", newSpanEndTimestamp.Sub(newSpanStartTimestamp))
			}

			span["start_timestamp"], _ = newSpanStartTimestamp.Round(7).Float64()
			span["timestamp"], _ = newSpanEndTimestamp.Round(7).Float64()

			// fmt.Printf("\n> updatetimestamps SPAN start_timestamp after %v (%T)", decimal.NewFromFloat(span["start_timestamp"].(float64)), span["start_timestamp"])
			// fmt.Printf("\n> updatetimestamps SPAN       timestamp after %v (%T)\n", decimal.NewFromFloat(span["timestamp"].(float64)), span["timestamp"])
		}
	}
}
