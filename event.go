package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/getsentry/sentry-go"
)

type TypeSwitch struct {
	Kind string `json:"type"`
}

type Event struct {
	TypeSwitch `json:"type"`
	Platform   string
	*Error
	*Transaction
	*DSN
}

const DEFAULT = "default"

func (event *Event) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &event.TypeSwitch); err != nil {
		return err
	}
	// may not need
	if event.Kind == "" {
		fmt.Println("> event.Kind", event.Kind)
		sentry.CaptureMessage("no event.Kind set")
		log.Fatal("no event.Kind set")
	}
	switch event.Kind {
	case ERROR:
		event.Error = &Error{}
		return json.Unmarshal(data, event.Error)
	case TRANSACTION:
		event.Transaction = &Transaction{}
		return json.Unmarshal(data, event.Transaction)
	case DEFAULT:
		event.Error = &Error{}
		return json.Unmarshal(data, event.Error)
	default:
		sentry.CaptureMessage("unrecognized type value " + event.Kind)
		return fmt.Errorf("unrecognized type value %q", event.Kind)
	}
}

func (event *Event) getPlatform() string {
	var platform string
	if event.Kind == TRANSACTION {
		platform = event.Transaction.Platform
	}
	if event.Kind == ERROR {
		platform = event.Error.Platform
	}
	if event.Kind == DEFAULT {
		platform = event.Error.Platform
	}
	if platform == "" {
		sentry.CaptureException(errors.New("no event platform set"))
		log.Fatalf("no event platform set")
	}
	return platform
}

func (event *Event) setDsn(dsn string) {
	event.DSN = NewDSN(dsn)
	if event.DSN == nil {
		sentry.CaptureException(errors.New("null DSN"))
		log.Fatal("null DSN")
	}
}

func (event *Event) setPlatform() {
	for _, platform := range platforms {
		if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == platform {
			event.Platform = platform
			break
		} else if event.Kind == TRANSACTION && event.Transaction.Platform == platform {
			event.Platform = platform
			break
		}
	}
	if event.Platform == "" {
		sentry.CaptureException(errors.New("event.Kind and Platform condition not found" + event.Kind))
		log.Fatalf("event Kind: %v and Platform: %v not recognized", event.Kind, event.Platform)
	}
}

// Undertaker adds the replay tag
func (e Event) undertake() {
	if e.Kind == ERROR || e.Kind == DEFAULT {
		if e.Error.Tags == nil {
			e.Error.Tags = make([][]string, 0)
		}
		// TODO if []Tags already has 'replay', then it gets duplicated in the array, but doesn't error in Sentry
		// tagItem := []string{"replay", "replay"}
		// e.Error.Tags = append(e.Error.Tags, tagItem)
	}
	if e.Kind == TRANSACTION {
		if e.Transaction.Tags == nil {
			e.Transaction.Tags = make([][]string, 0)
		}
		// tagItem := []string{"replay", "replay"}
		// e.Transaction.Tags = append(e.Transaction.Tags, tagItem)
	}
}
