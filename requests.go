package main

type Requests struct {
	events []Event
}

func (r *Requests) send() {
	for _, event := range r.events {
		request := NewRequest(event)
		request.send()
	}
}
