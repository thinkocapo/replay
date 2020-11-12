package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

// same eventId cannot be accepted twice by Sentry
func eventId(body map[string]interface{}) map[string]interface{} {
	if _, ok := body["event_id"]; !ok {
		log.Print("no event_id on object from DB")
	}
	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
	body["event_id"] = uuid4
	fmt.Println("\n> event_id updated", body["event_id"])
	return body
}

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
}

func sentAt(envelopeItems []interface{}) []interface{} {
	for _, item := range envelopeItems {
		sentAt := item.(map[string]interface{})["sent_at"]
		if sentAt != nil {
			item.(map[string]interface{})["sent_at"] = time.Now().UTC()
		}
	}
	return envelopeItems
}

func envelopeReleases(envelopeItems []interface{}, platform string, kind string) []interface{} {
	for _, item := range envelopeItems {
		currentRelease := item.(map[string]interface{})["release"]
		if currentRelease != nil {
			// "cannot call non-function release"
			// item = release(item.(map[string]interface{}))
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
			item.(map[string]interface{})["release"] = release
		}
	}
	return envelopeItems
}

// CalVer https://calver.org/
func release(body map[string]interface{}) map[string]interface{} {
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
	body["release"] = release
	return body
}

func user(body map[string]interface{}) map[string]interface{} {
	body["user"] = make(map[string]interface{})
	user := body["user"].(map[string]interface{})
	user["email"] = createUser()
	return body
}

func users(envelopeItems []interface{}) []interface{} {
	for _, item := range envelopeItems {
		user := item.(map[string]interface{})["user"]
		if user != nil {
			user.(map[string]interface{})["email"] = createUser()
		}
	}
	return envelopeItems
}

func createUser() string {
	rand.Seed(time.Now().UnixNano())
	alpha := strings.Split("abcdefghijklmnopqrstuvwxyz", "")[rand.Intn(9)]
	var alphanumeric string
	for i := 0; i < 3; i++ {
		alphanumeric += strings.Split("abcdefghijklmnopqrstuvwxyz0123456789", "")[rand.Intn(35)]
	}
	return fmt.Sprint(alpha, alphanumeric, "@yahoo.com")
}

func undertake(body map[string]interface{}) {
	if body["tags"] == nil {
		body["tags"] = make(map[string]interface{})
	}
	tags := body["tags"].(map[string]interface{})
	tags["undertaker"] = "h4ckweek"
}

// Python Transactions have "length". Remove it or else rejected.
func removeLengthField(items []interface{}) []interface{} {
	for _, item := range items {
		delete(item.(map[string]interface{}), "length")
	}
	return items
}

func getEnvelopeTraceIds(items []interface{}) {
	for _, item := range items {
		contexts := item.(map[string]interface{})["contexts"]

		if contexts != nil {
			if _, found := contexts.(map[string]interface{})["trace"]; found {
				trace := contexts.(map[string]interface{})["trace"]
				trace_id := trace.(map[string]interface{})["trace_id"].(string)
				if trace_id != "" {
					//traceIdMap[trace_id] = append(traceIdMap[trace_id], item)
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
}

// Runs after all transactions (envelopes) have been iterated through.
func setEnvelopeTraceIds(requests []Transport) {
	fmt.Println("\n> setEnvelopeTraceIds <", traceIds)

	for _, TRACE_ID := range traceIds {
		//fmt.Println("\n> TRACE_ID", TRACE_ID)

		var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
		NEW_TRACE_ID := uuid4

		for _, transport := range requests {
			//fmt.Println("> TRANSPORT", idx, transport.kind, transport.platform)

			if transport.kind == "error" {
				contexts := transport.bodyError["contexts"]
				if contexts != nil {
					trace := contexts.(map[string]interface{})["trace"]
					if TRACE_ID == trace.(map[string]interface{})["trace_id"] {
						// fmt.Println("\n> MATCHED Error trace_id BEFORE", trace.(map[string]interface{})["trace_id"])
						trace.(map[string]interface{})["trace_id"] = NEW_TRACE_ID
						// fmt.Println("> MATCHED Error trace_id AFTER", transport.bodyError["contexts"].(map[string]interface{})["trace"].(map[string]interface{})["trace_id"].(string))
					}
				}
			}
			if transport.kind == "transaction" {
				for _, item := range transport.envelopeItems {
					contexts := item.(map[string]interface{})["contexts"]
					if contexts != nil {
						trace := contexts.(map[string]interface{})["trace"]
						if TRACE_ID == trace.(map[string]interface{})["trace_id"] {
							trace.(map[string]interface{})["trace_id"] = NEW_TRACE_ID
							//fmt.Println(">   MATCHED Transaction trace_id AFTER", item.(map[string]interface{})["contexts"].(map[string]interface{})["trace"].(map[string]interface{})["trace_id"].(string))
							if _, found := item.(map[string]interface{})["spans"]; found {
								spans := item.(map[string]interface{})["spans"]
								if len(spans.([]interface{})) > 0 {
									for _, value := range spans.([]interface{}) {
										// fmt.Println("\n> BEFORE ", value.(map[string]interface{})["trace_id"])
										value.(map[string]interface{})["trace_id"] = NEW_TRACE_ID
										//fmt.Println(">   SPAN Transaction trace_id AFTER", item.(map[string]interface{})["spans"].([]interface{})[0].(map[string]interface{})["trace_id"])
									}
								}
							}
						}
					}
				}
			}
		}

	}
}
