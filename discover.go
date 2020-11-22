package main

import "encoding/json"

type Discover struct {
	// TODO
}

func (d *Discover) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, d)
	return nil
}
