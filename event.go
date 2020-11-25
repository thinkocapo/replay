package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

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
	if event.Kind == "" {
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

func (event *Event) setDsn() {
	if event.Kind == TRANSACTION && event.Transaction.Platform == JAVASCRIPT {
		event.DSN = NewDSN(os.Getenv("DSN_JAVASCRIPT_SAAS"))
		event.Platform = JAVASCRIPT
	} else if event.Kind == TRANSACTION && event.Transaction.Platform == PYTHON {
		event.DSN = NewDSN(os.Getenv("DSN_PYTHON_SAAS"))
		event.Platform = PYTHON
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == JAVASCRIPT {
		event.DSN = NewDSN(os.Getenv("DSN_JAVASCRIPT_SAAS"))
		event.Platform = JAVASCRIPT
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == PYTHON {
		event.DSN = NewDSN(os.Getenv("DSN_PYTHON_SAAS"))
		event.Platform = PYTHON
	} else {
		sentry.CaptureException(errors.New("event.Kind and Type condition not found" + event.Kind))
		log.Fatal("event.Kind and type not recognized " + event.Kind)
	}
}

//TODO - Destinations
// only do if event.platform matches what platform:DSN is intended for
// func (event *Event) setDsn(fullurl) {
// 	dsn := NewDSN(fullurl)
// }
// or
// func (event *Event) setDsn(dsn DSN) {

// 	for envvarKey, envarValue in envarPairs {
// 		switch envarKeys:
// 		case DSN_JAVASCRIPT
// 			// should be Private
// 			// for python_destinations from YAML:
// 				// destination()
// 		case DSN_PYTHON
// 	}

// the fact that it does each destination one-by-one, gives each a little bit of a pause, like a Sleep Timeout, so no need to code a short Sleep Timeout
