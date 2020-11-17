package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

func createUser() string {
	rand.Seed(time.Now().UnixNano())
	alpha := strings.Split("abcdefghijklmnopqrstuvwxyz", "")[rand.Intn(9)]
	var alphanumeric string
	for i := 0; i < 3; i++ {
		alphanumeric += strings.Split("abcdefghijklmnopqrstuvwxyz0123456789", "")[rand.Intn(35)]
	}
	return fmt.Sprint(alpha, alphanumeric, "@yahoo.com")
}

func getTraceIds(events []Event) {
	// var traceIds []string
	for _, event := range events {
		var contexts map[string]interface{}
		if event.Kind == "error" {
			contexts = event.Error.Contexts
		}
		if event.Kind == "transaction" {
			contexts = event.Transaction.Contexts
		}
		if contexts != nil {
			// fmt.Println("> getTraceIds context != nil")
			if _, found := contexts["trace"]; found {
				trace := contexts["trace"]
				trace_id := trace.(map[string]interface{})["trace_id"].(string)
				if trace_id != "" {
					// fmt.Println("> getTraceIds trace_id != nil")
					matched := false
					for _, value := range traceIds {
						if trace_id == value {
							matched = true
						}
					}
					if !matched {
						traceIds = append(traceIds, trace_id)
					}
				}
			}
		}
	}
	fmt.Println("> getTraceids traceIds", traceIds)
}

func undertake(body map[string]interface{}) {
	if body["tags"] == nil {
		body["tags"] = make(map[string]interface{})
	}
	tags := body["tags"].(map[string]interface{})
	tags["undertaker"] = "h4ckweek"
}

func updateTraceIds(events []Event) {
	for _, TRACE_ID := range traceIds {
		var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
		NEW_TRACE_ID := uuid4

		for _, event := range events {
			if event.Kind == "error" {
				contexts := event.Error.Contexts
				if contexts != nil {
					trace := contexts["trace"]
					if TRACE_ID == trace.(map[string]interface{})["trace_id"] {
						// fmt.Println("\n> MATCHED Error trace_id BEFORE", trace.(map[string]interface{})["trace_id"])
						trace.(map[string]interface{})["trace_id"] = NEW_TRACE_ID
						// fmt.Println("> MATCHED Error trace_id AFTER", transport.bodyError["contexts"].(map[string]interface{})["trace"].(map[string]interface{})["trace_id"].(string))
					}
				}
			}
			if event.Kind == "transaction" {
				contexts := event.Transaction.Contexts
				if contexts != nil {
					trace := contexts["trace"]
					if TRACE_ID == trace.(map[string]interface{})["trace_id"] {
						trace.(map[string]interface{})["trace_id"] = NEW_TRACE_ID
						//fmt.Println(">   MATCHED Transaction trace_id AFTER", item.(map[string]interface{})["contexts"].(map[string]interface{})["trace"].(map[string]interface{})["trace_id"].(string))

						// TODO should check if 'Spans' field exists. it may have been set to 0 if nothing was unmarshal'd to it
						if len(event.Transaction.Spans) > 0 {
							spans := event.Transaction.Spans
							// TODO then should check if length is gt 0
							// if len(spans.([]interface{})) > 0 {
							for _, value := range spans {
								// fmt.Println("\n> SPAN Transaction trace_id BEFORE ", value["trace_id"])
								value["trace_id"] = NEW_TRACE_ID
								// fmt.Println("> SPAN Transaction trace_id AFTER", event.Transaction.Spans[0]["trace_id"])
							}
							// }
						}
					}
				}
			}
		}

	}
}
