package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"math/rand"
	"github.com/google/uuid"
	"strings"
	"time"
)

// same eventId cannot be accepted twice by Sentry
func eventId(body map[string]interface{}) map[string]interface{} {
	if _, ok := body["event_id"]; !ok {
		log.Print("no event_id on object from DB")
	}
	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
	body["event_id"] = uuid4
	fmt.Println("> event_id updated", body["event_id"])
	return body
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
	fmt.Println("> release", body["release"])
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

// Python Transactions have "length"
func removeLengthField(items []interface{}) []interface{} {
	for _, item := range items {
		delete(item.(map[string]interface{}), "length")
	}
	return items
}