package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Error struct {
	EventId   string                 `json:"event_id"`
	Release   string                 `json:"release"`
	User      map[string]interface{} `json:"user"`
	Timestamp int64                  `json:"timestamp"`
	// Type      string                 `json:"type"`
}

func (e Error) eventId() Error {
	// if _, ok := e.EventId; !ok {
	// 	log.Print("no event_id on object from DB")
	// }
	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
	e.EventId = uuid4
	fmt.Println("\n> event_id updated", e.EventId)
	return e
}

// CalVer https://calver.org/
func (e Error) release() Error {
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
	return e
}

func (e Error) user() Error {
	e.User = make(map[string]interface{})
	user := e.User //.(map[string]interface{})
	user["email"] = createUser()
	return e
}

func (e Error) setTimestamp() Error {
	e.Timestamp = time.Now().Unix()
	return e
}
