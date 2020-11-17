package main

import (
	"encoding/json"
	"fmt"
)

type TypeSwitch struct {
	Kind string `json:"type"`
}

type Event struct {
	TypeSwitch `json:"type"`
	*Error
	*Transaction
}

func (event *Event) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &event.TypeSwitch); err != nil {
		return err
	}
	switch event.Kind {
	case "error":
		event.Error = &Error{}
		return json.Unmarshal(data, event.Error)
	case "transaction":
		event.Transaction = &Transaction{}
		return json.Unmarshal(data, event.Transaction)
	default:
		return fmt.Errorf("unrecognized type value %q", event.Kind)
	}
}
