package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type TypeSwitch struct {
	Kind string `json:"type"`
}

type Event struct {
	TypeSwitch `json:"type"`
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
		fmt.Println(" * * * * ** * * NOTHING....", event)
		// TODO - sentry,
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
		event.Error = &Error{}
		return json.Unmarshal(data, event.Error)
		// sentry.CaptureMessage("unrecognized type value " + event.Kind)
		// return fmt.Errorf("unrecognized type value %q", event.Kind)
	}
	return nil // TODO test
}

func (event *Event) setDsn() {
	if event.Kind == TRANSACTION && event.Transaction.Platform == JAVASCRIPT {
		event.DSN = NewDSN(os.Getenv("DSN_JAVASCRIPT_SAAS"))
	} else if event.Kind == TRANSACTION && event.Transaction.Platform == PYTHON {
		event.DSN = NewDSN(os.Getenv("DSN_PYTHON_SAAS"))
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == JAVASCRIPT {
		event.DSN = NewDSN(os.Getenv("DSN_JAVASCRIPT_SAAS"))
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == PYTHON {
		event.DSN = NewDSN(os.Getenv("DSN_PYTHON_SAAS"))
	} else {
		fmt.Println("XXXXXXXXXX", event.Kind)
	}
	// TODO 7:41p add else condition
}
