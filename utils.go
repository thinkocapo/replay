package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

func createUser() string {
	rand.Seed(time.Now().UnixNano())
	alpha := strings.Split("abcdefghijklmnopqrstuvwxyz", "")[rand.Intn(9)]
	var alphanumeric string
	for i := 0; i < 3; i++ {
		alphanumeric += strings.Split("abcdefghijklmnopqrstuvwxyz0123456789", "")[rand.Intn(35)]
	}
	return fmt.Sprintf("%v%v@yahoo.com", alpha, alphanumeric)
}

func getTraceIds(events []Event) {
	for _, event := range events {
		var contexts map[string]interface{}
		if event.Kind == ERROR {
			contexts = event.Error.Contexts
		}
		if event.Kind == TRANSACTION {
			contexts = event.Transaction.Contexts
		}
		if contexts != nil {
			if _, found := contexts["trace"]; found {
				trace := contexts["trace"]
				trace_id := trace.(map[string]interface{})["trace_id"].(string)
				if trace_id != "" {
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

func initializeSentry() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         os.Getenv("SENTRY"),
		Environment: os.Getenv("ENVIRONMENT"),
		Release:     time.Now().Month().String(),
		// Debug:       true,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	if hostName, _ := os.Hostname(); hostName != "" {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{Username: hostName})
		})
	}
	defer sentry.Flush(2 * time.Second)
}

type Config struct {
	Sources      []string
	Destinations struct {
		Javascript []string `yaml:"javascript"`
		Python     []string `yaml:"python"`
	}
}

func parseConfig() {
	filename := "config.yml"
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		sentry.CaptureException(err)
		panic(err)
	}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		sentry.CaptureException(err)
		panic(err)
	}
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
			if event.Kind == ERROR {
				contexts := event.Error.Contexts
				if contexts != nil {
					trace := contexts["trace"]
					if trace != nil { // need this or else kind:default's error out

						if TRACE_ID == trace.(map[string]interface{})["trace_id"] {
							// fmt.Println("\n> MATCHED Error trace_id BEFORE", trace.(map[string]interface{})["trace_id"])
							trace.(map[string]interface{})["trace_id"] = NEW_TRACE_ID
							// fmt.Println("> MATCHED Error trace_id AFTER", transport.bodyError["contexts"].(map[string]interface{})["trace"].(map[string]interface{})["trace_id"].(string))
						}
					}
				}
			}
			if event.Kind == TRANSACTION {
				contexts := event.Transaction.Contexts
				if contexts != nil {
					trace := contexts["trace"]
					if TRACE_ID == trace.(map[string]interface{})["trace_id"] {
						trace.(map[string]interface{})["trace_id"] = NEW_TRACE_ID
						//fmt.Println(">   MATCHED Transaction trace_id AFTER", item.(map[string]interface{})["contexts"].(map[string]interface{})["trace"].(map[string]interface{})["trace_id"].(string))

						// should check if 'Spans' field exists. it may have been set to 0 if nothing was unmarshal'd to it
						if len(event.Transaction.Spans) > 0 {
							spans := event.Transaction.Spans
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
