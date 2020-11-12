package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Error struct {
	EventId   string                 `json:"event_id"`
	Release   string                 `json:"release"`
	User      map[string]interface{} `json:"user"`
	Timestamp float64                `json:"timestamp"`
	// Type      string                 `json:"type"`
}

func (e Error) eventId() Error {
	// if _, ok := e.EventId; !ok {
	// 	log.Print("no event_id on object from DB")
	// }
	var uuid4 = strings.ReplaceAll(uuid.New().String(), "-", "")
	e.EventId = uuid4
	// fmt.Println("\n> event_id updated", e.EventId)
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
	timestamp := fmt.Sprint(time.Now().Unix())
	timestampDecimal, err1 := decimal.NewFromString(timestamp[:10] + "." + timestamp[10:])
	fmt.Print("> timestampDecimal", timestampDecimal)
	if err1 != nil {
		log.Fatal(err1)
	}
	timestampFinal, err2 := timestampDecimal.Round(7).Float64()
	if err2 == false {
		log.Fatal(err2)
	}
	e.Timestamp = timestampFinal
	return e
}
