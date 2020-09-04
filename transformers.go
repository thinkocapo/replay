package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
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
			item.(map[string]interface{})["event_id"] = uuid4
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

		/*
			contexts := item.(map[string]interface{})["contexts"]
			if (contexts != nil) {
				// fmt.Println("\n\n > > THIS HAS CONTEXT")
				// fmt.Println("\n > > contexts RELEASE", contexts.(map[string]interface{})["release"])

				// trace := contexts.(map[string]interface{})["trace"]
				// fmt.Println("\n > > trace RELEASE", trace.(map[string]interface{})["release"])

				fmt.Println("\n > > Release BEFORE", item.(map[string]interface{})["contexts"].(map[string]interface{})["trace"].(map[string]interface{})["release"])
				item.(map[string]interface{})["contexts"].(map[string]interface{})["trace"].(map[string]interface{})["release"] = "619"
				fmt.Println("\n > > Release AFTER", item.(map[string]interface{})["contexts"].(map[string]interface{})["trace"].(map[string]interface{})["release"])

				// NO because nested too far
				// item = release(trace.(map[string]interface{}))
			}
		*/
		// release := trace.(map[string]interface{})["release"]
	}
	return envelopeItems
}

// CalVer-lite
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
	if body["user"] == nil {
		body["user"] = make(map[string]interface{})
		user := body["user"].(map[string]interface{})
		rand.Seed(time.Now().UnixNano())
		alpha := strings.Split("abcdefghijklmnopqrstuvwxyz", "")[rand.Intn(9)]
		var alphanumeric string
		for i := 0; i < 3; i++ {
			alphanumeric += strings.Split("abcdefghijklmnopqrstuvwxyz0123456789", "")[rand.Intn(35)]
		}
		user["email"] = fmt.Sprint(alpha, alphanumeric, "@yahoo.com")
	}
	return body
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

// TODO could put this to decodeEnvelope? and return it to event-to-sentry. or reference this func from there
//item.context.trace.traceId
func getEnvelopeTraceIds(items []interface{}) {
	for _, item := range items {
		contexts := item.(map[string]interface{})["contexts"]

		// fmt.Println("> getEnvelopeTraceIds", contexts)

		if contexts != nil {
			if _, found := contexts.(map[string]interface{})["trace"]; found { // if value, found :=
				trace := contexts.(map[string]interface{})["trace"]
				trace_id := trace.(map[string]interface{})["trace_id"].(string)
				// fmt.Printf("> VICTORY...trace_id BEFORE%v\n", trace_id)
				if trace_id != "" {
					// timestamp := item.(map[string]interface{})["timestamp"].(string)
					traceIdMap[trace_id] = append(traceIdMap[trace_id], item)
					// fmt.Println("\n VICTOR trace_id ", trace_id)
				}
			}
		}

		// MP
		// contexts := item.(map[string]interface{})["contexts"]
		// if (contexts != nil) {
		// trace := contexts.(map[string]interface{})["trace"]

		// 	trace := contexts.(map[string]interface{})["trace"].(map[string]interface{})
		// 	trace_id := trace["trace_id"].(string)
		// 	fmt.Println("> trace_id BEFORE1", trace_id)
		// 	fmt.Println("\n > trace_id", trace_id)

		// 	if trace_id != "" {
		// 		// timestamp := item.(map[string]interface{})["timestamp"].(string)
		// 		fmt.Println("\n trace_id ", trace_id)
		// 		traceIdMap[trace_id] = append(traceIdMap[trace_id], item)
		// 	}
		// }

	}
}

// Runs after all transactions (envelopes) have been iterated through.
func setEnvelopeTraceIds() {
	// compiles but not needed
	// for _, item := range traceIdMap[trace_id] {

	// self.trace_id = trace_id or uuid.uuid4().hex
	// self.span_id = span_id or uuid.uuid4().hex[16:]

	// if trace_id is not None:
	// trace_id = "{:032x}".format(int(trace_id, 16))
	// if span_id is not None:
	// span_id = "{:016x}".format(int(span_id, 16))

	for currentTrace_id, items := range traceIdMap {
		var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
		newTraceId := uuid4

		fmt.Println("\n> setEnvelopeTraceIds trace_id", currentTrace_id)
		for _, item := range items {
			// contexts := item.(map[string]interface{})
			contexts := item.(map[string]interface{})["contexts"]
			trace := contexts.(map[string]interface{})["trace"]
			// trace := contexts.(map[string]interface{})["trace"]

			// fmt.Println("> trace_id BEFORE2", trace_id)
			trace.(map[string]interface{})["trace_id"] = newTraceId
			fmt.Println("> trace_id AFTER", item.(map[string]interface{})["contexts"].(map[string]interface{})["trace"])
			// fmt.Println("> trace_id AFTER", item.(map[string]interface{})["context"].(map[string]interface{})["trace"].(map[string]interface{}))
		}
	}
}
