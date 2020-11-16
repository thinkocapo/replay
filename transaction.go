package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	EventId   string                 `json:"event_id"`
	Release   string                 `json:"release"`
	User      map[string]interface{} `json:"user"`
	Timestamp float64                `json:"timestamp"`
	Type      string                 `json:"type"`
	Platform  string                 `json:"platform"`

	Project int        `json:"project"`
	Message string     `json:"message"`
	Tags    [][]string `json:"tags"`

	Breadcrumbs map[string]interface{} `json:"breadcrumbs"`
	Contexts    map[string]interface{} `json:"contexts"`
	Culprit     string                 `json:"culprit"`

	Environment string `json:"environment"`

	Extra map[string]interface{} `json:"extra"`

	Grouping_config map[string]interface{} `json:"grouping_config"` // nothing new but also no processing error warnings

	Key_id string `json:"key_id"`
	Level  string `json:"level"`
	Logger string `json:"logger"`

	Metadata map[string]interface{} `json:"metadata"`
	Received float64                `json:"received"`
	Request  map[string]interface{} `json:"request"`

	Sdk map[string]interface{} `json:"sdk"`

	Version string `json:"version"`

	Spans []map[string]interface{} `json:"spans"`

	Start_timestamp float64 `json:"start_timestamp"`
	Transaction     string  `json:"transaction"`

	// TODO measurements, extra
}

func (t *Transaction) eventId() {

	// if _, ok := t.EventId; !ok {
	// 	log.Print("no event_id on object from DB")
	// }

	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
	t.EventId = uuid4
	// fmt.Println("\n******** event_id updated *********", t.EventId)

	// LOOKING like only 1 event_id in the *transctino.json file ;)
	/*
		func eventIds(envelopeItems []interface{}) []interface{} {
			var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
			for _, item := range envelopeItems {
				eventId := item.(map[string]interface{})["event_id"]
				if eventId != nil {
					fmt.Println("\n> event_id eventIds", uuid4)
					item.(map[string]interface{})["event_id"] = uuid4
				}
			}
			return envelopeItems
		}*/

	// 1 find where in the event-1-*transaction.json's there are eventId's
	// 2 update each

}

// setting here, and tag may get value from it?
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
	/*unixTimestamp := fmt.Sprint(time.Now().Unix())
	decimalTimestamp, err1 := decimal.NewFromString(unixTimestamp[:10] + "." + unixTimestamp[10:])
	// fmt.Print("> decimalTimestamp\n", decimalTimestamp)
	if err1 != nil {
		log.Fatal(err1)
	}
	timestamp, err2 := decimalTimestamp.Round(7).Float64()
	if err2 == false {
		log.Fatal(err2)
	}
	t.Timestamp = timestamp*/

	// fmt.Printf("\n> updateTimestamps PARENT start_timestamp before %v (%T) ", t.Start_timestamp, t.Start_timestamp)
	// fmt.Printf("\n> updateTimestamps PARENT       timestamp before %v (%T) \n", t.Timestamp, t.Timestamp)

	if t.Timestamp != 0 && t.Start_timestamp != 0 {
		var parentStartTimestamp, parentEndTimestamp decimal.Decimal
		if t.Platform == "python" {
			fmt.Print("**** I'M PYTHON ****")
			// parentStart, _ := time.Parse(time.RFC3339Nano, fmt.Sprintf("%f", t.Start_timestamp))
			// parentEnd, _ := time.Parse(time.RFC3339Nano, fmt.Sprintf("%f", t.Timestamp))
			// parentStartTime := fmt.Sprint(parentStart.UnixNano())
			// parentEndTime := fmt.Sprint(parentEnd.UnixNano())
			// parentStartTimestamp, _ = decimal.NewFromString(parentStartTime[:10] + "." + parentStartTime[10:])
			// parentEndTimestamp, _ = decimal.NewFromString(parentEndTime[:10] + "." + parentEndTime[10:])

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
			st := fmt.Sprintf("%f", span["start_timestamp"].(float64))
			s := fmt.Sprintf("%f", span["timestamp"].(float64))
			fmt.Println("> st ", st)
			fmt.Println("> s ", s)

			if t.Platform == "python" {
				// spanStart, _ := time.Parse(time.RFC3339Nano, fmt.Sprintf("%f", span["start_timestamp"].(float64)))
				// spanEnd, _ := time.Parse(time.RFC3339Nano, fmt.Sprintf("%f", span["timestamp"].(float64)))
				// fmt.Println("> spanStart", spanStart)
				// fmt.Println("> spanEnd", spanEnd)
				// spanStartTime := fmt.Sprint(spanStart.UnixNano())
				// spanEndTime := fmt.Sprint(spanEnd.UnixNano())
				// spanStartTimestamp, _ = decimal.NewFromString(spanStartTime[:10] + "." + spanStartTime[10:])
				// spanEndTimestamp, _ = decimal.NewFromString(spanEndTime[:10] + "." + spanEndTime[10:])

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

			fmt.Printf("\n> updatetimestamps SPAN start_timestamp after %v (%T)", decimal.NewFromFloat(span["start_timestamp"].(float64)), span["start_timestamp"])
			fmt.Printf("\n> updatetimestamps SPAN       timestamp after %v (%T)\n", decimal.NewFromFloat(span["timestamp"].(float64)), span["timestamp"])
		}
	}
}

func (t *Transaction) traceIds() {

}

// not seeing 'sent_at sentAt' property on post-ingest transaction (it was on the pre-ingest tx), so not defining func (t *Transaction) sentAt()
// not seeing 'length        ' property on post-ingest transaction (it was on the pre-ingest tx), so not defining func (t *Transaction) sentAt()
