package main

type Transaction struct {
	EventId   string                 `json:"event_id"`
	Release   string                 `json:"release"`
	User      map[string]interface{} `json:"user"`
	Timestamp int64                  `json:"timestamp"`
	// Type      string                 `json:"type"`
}
