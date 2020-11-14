package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

/*
Post-Ingestion from https://sentry.io/api/0/projects/<org>/<project>/events/<event_id>/json/
skip _meta, _metrics, errors
*/
type Error struct {
	EventId   string                 `json:"event_id"`
	Release   string                 `json:"release"`
	User      map[string]interface{} `json:"user"`
	Timestamp float64                `json:"timestamp"`
	Type      string                 `json:"type"`
	Platform  string                 `json:"platform"`

	Project int `json:"project"`
	// Dist     string `json:"dist"`
	Message string `json:"message"`
	// Datetime string `json:"datetime"`
	Tags            [][]string             `json:"tags"`
	Breadcrumbs     map[string]interface{} `json:"breadcrumbs"`
	Contexts        map[string]interface{} `json:"contexts"`
	Culprit         string                 `json:"culprit"`
	Environment     string                 `json:"environment"`
	Exception       map[string]interface{} `json:"exception"`
	Fingerprint     []string               `json:"fingerprint"`     // nothing new but also no processing error warnings
	Grouping_config map[string]interface{} `json:"grouping_config"` // nothing new but also no processing error warnings
	Hashes          []string               `json:"hashes"`          // nothing new but also no processing error warnings
	Key_id          string                 `json:"key_id"`
	Level           string                 `json:"level"`
	Location        string                 `json:"location"`
	Logger          string                 `json:"logger"`
}

func (e *Error) eventId() {
	// if _, ok := e.EventId; !ok {
	// 	log.Print("no event_id on object from DB")
	// }
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

func (e *Error) user() {
	e.User = make(map[string]interface{})
	user := e.User //.(map[string]interface{})
	user["email"] = createUser()
}

func (e *Error) setTimestamp() {
	timestamp := fmt.Sprint(time.Now().Unix())
	timestampDecimal, err1 := decimal.NewFromString(timestamp[:10] + "." + timestamp[10:])
	fmt.Print("> timestampDecimal\n", timestampDecimal)
	if err1 != nil {
		log.Fatal(err1)
	}
	timestampFinal, err2 := timestampDecimal.Round(7).Float64()
	if err2 == false {
		log.Fatal(err2)
	}
	e.Timestamp = timestampFinal
}
