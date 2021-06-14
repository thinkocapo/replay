package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

/*
Post-Ingestion from https://sentry.io/api/0/projects/<org>/<project>/events/<event_id>/json/
Either don't sound needed or give 'Discarded unknown attribute'
skip: dist, datetime, _meta, _metrics, errors, location, title
*/
type Error struct {
	EventId   string                 `json:"event_id"`
	Release   string                 `json:"release"`
	User      map[string]interface{} `json:"user"`
	Timestamp float64                `json:"timestamp"`
	Type      string                 `json:"type"`
	Platform  string                 `json:"platform"`

	Project     int                    `json:"project"`
	Message     string                 `json:"message"`
	Tags        [][]string             `json:"tags"`
	Breadcrumbs map[string]interface{} `json:"breadcrumbs"`
	Contexts    map[string]interface{} `json:"contexts"`
	Culprit     string                 `json:"culprit"`
	// Environment     string                 `json:"environment"`
	Exception       map[string]interface{} `json:"exception"`
	Fingerprint     []string               `json:"fingerprint"`
	Grouping_config map[string]interface{} `json:"grouping_config"`
	Hashes          []string               `json:"hashes"`
	Key_id          string                 `json:"key_id"`
	Level           string                 `json:"level"`
	Logger          string                 `json:"logger"`
	Metadata        map[string]interface{} `json:"metadata"`
	Received        float64                `json:"received"`
	Request         map[string]interface{} `json:"request"`
	Sdk             map[string]interface{} `json:"sdk"`
	Version         string                 `json:"version"`
	Extra           map[string]interface{} `json:"extra"`
	Modules         map[string]interface{} `json:"modules"`
	// Threads and DebugMeta were added for cocoa support
	Threads    map[string]interface{} `json:"threads"`
	Debug_meta map[string]interface{} `json:"debug_meta"`
}

const ERROR = "error"

func (e *Error) eventId() {
	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
	e.EventId = uuid4
}

// CalVer https://calver.org/
func (e *Error) release() {
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
	e.Release = release
}

// TODO sync same user across all errors+tx's in the dataset
func (e *Error) user() {
	if e.User == nil {
		e.User = make(map[string]interface{})
	}
	user := e.User
	user["email"] = createUser()
}

/*
PYTHON timestamp format is 2020-06-06T04:54:56.636664Z RFC3339Nano
JAVASCRIPT timestamp format is 1591419091.4805 to 1591419092.000035
PARENT TRACE - Adjust the parentDifference/spanDifference between .01 and .2 (1% and 20% difference) so the 'end timestamp's always shift the same amount (no gaps at the end)
TRANSACTIONS. body.contexts.trace.span_id is the Parent Trace. start/end here is same as the sdk's start_timestamp/timestamp, and start_timestamp is only present in transactions
To see a full span `firstSpan := body["spans"].([]interface{})[0].(map[string]interface{})``
7 decimal places as the range sent by sdk's is 4 to 7
https://www.epochconverter.com/
Float form is 1.5914674155654302e+09
*/
func (e *Error) timestamp() {
	unixTimestamp := fmt.Sprint(time.Now().Unix())
	decimalTimestamp, err1 := decimal.NewFromString(unixTimestamp[:10] + "." + unixTimestamp[10:])
	if err1 != nil {
		sentry.CaptureException(err1)
		log.Fatal(err1)
	}
	timestamp, err2 := decimalTimestamp.Round(7).Float64()
	if err2 == false {
		// sentry.CaptureException(err2)
		log.Fatal(err2)
	}
	e.Timestamp = timestamp
}
