package main

type Requests struct {
	events []EventJson
}

func (r *Requests) send() {
	for _, event := range r.events {
		request := NewRequest(event)
		request.send()
	}
}

// Don't need because can `requests := Requests{events}``
// func NewRequests(events []EventJson) Requests {
// 	return Requests{events}
// }
