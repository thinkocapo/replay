package main

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
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
	// Title           string  `json:"title"` 'discarded attribute'
	Transaction string `json:"transaction"`
}

func (t *Transaction) eventId() {

	// if _, ok := t.EventId; !ok {
	// 	log.Print("no event_id on object from DB")
	// }

	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
	t.EventId = uuid4
	fmt.Println("\n******** event_id updated *********", t.EventId)

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
