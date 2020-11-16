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
