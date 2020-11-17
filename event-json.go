package main

import (
	"encoding/json"
	"fmt"
)

type TypeSwitch struct {
	Kind string `json:"type"`
}

// TODO Event instead of EventJson?
type EventJson struct {
	TypeSwitch `json:"type"`
	*Error
	*Transaction
}

func (eventJson *EventJson) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &eventJson.TypeSwitch); err != nil {
		return err
	}
	switch eventJson.Kind {
	case "error":
		eventJson.Error = &Error{}
		return json.Unmarshal(data, eventJson.Error)
	case "transaction":
		eventJson.Transaction = &Transaction{}
		return json.Unmarshal(data, eventJson.Transaction)
	default:
		return fmt.Errorf("unrecognized type value %q", eventJson.Kind)
	}
}
