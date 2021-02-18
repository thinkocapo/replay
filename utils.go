package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/user"
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
	// fmt.Println("> getTraceids traceIds", traceIds)
}

// Capture errors if they occur during Replay's execution
func initializeSentry() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         config.SentryJobMonitor,
		Environment: config.Environment,
		Release:     time.Now().Month().String(),
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	if user, _ := user.Current(); user != nil {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{Username: user.Username, IPAddress: ip()})
		})
	}
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("demo-automation", "replay")
	})
	defer sentry.Flush(2 * time.Second)
}

func ip() string {
	url := "https://api.ipify.org?format=text"
	resp, err := http.Get(url)
	if err != nil {
		sentry.CaptureException(err)
		panic(err)
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		sentry.CaptureException(err)
		panic(err)
	}
	return string(ip)
}

type Config struct {
	Environment                  string
	SentryJobMonitor             string
	Bucket                       string
	GoogleApplicationCredentials string
	SentryAuthToken              string
	Sources                      []string
	Destinations                 map[string][]string
	Skip                         string
}

func parseYamlConfig() {
	filename := "config.yaml"
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

	var msg string

	// Must use job monitor if powering via DemoAutomation (scheduled job) in GCP
	if config.Environment == "production" && config.SentryJobMonitor == "" {
		msg = "no sentry job monitor dsn set"
	}
	if config.Bucket == "" {
		msg = "no bucket set"
	}
	if config.GoogleApplicationCredentials == "" {
		msg = "no app credentials set"
	} else {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", config.GoogleApplicationCredentials)
	}
	if len(config.Sources) == 0 {
		msg = "No sources defined"
	}
	// AuthToken is only if using DiscoverAPI EventsAPI in demo_automation.go, which is default disabled.
	// if config.SentryAuthToken == "" {
	// 	msg = "no auth token"
	// }
	// if config.Skip == "" {
	// 	msg = "no skip list provided"
	// }
	if msg != "" {
		sentry.CaptureException(errors.New(msg))
		log.Fatal(msg)
	}
}

func prettyPrint(v interface{}) {
	pp, _ := json.MarshalIndent(v, "", "  ")
	fmt.Print(string(pp))
}

func print(arg1 string, arg2 string) {
	fmt.Println(arg1, arg2)
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
