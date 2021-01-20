package main

import (
	"fmt"

	"github.com/getsentry/sentry-go"
)

type Requests struct {
	events []Event
}

// Doing each destination one-by-one, gives each org a rest before its API is called again, so don't insert a short Sleep Timeout yet
func (r *Requests) send() {
	for _, event := range r.events {
		// EVAL check if Destinations array is empty, or do during an init somewhere
		switch event.Platform {
		case JAVASCRIPT:
			for _, dsn := range config.Destinations.Javascript {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case PYTHON:
			for _, dsn := range config.Destinations.Python {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case JAVA:
			for _, dsn := range config.Destinations.Java {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case RUBY:
			for _, dsn := range config.Destinations.Ruby {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case GO:
			for _, dsn := range config.Destinations.Go {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case PHP:
			for _, dsn := range config.Destinations.Php {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case NODE:
			for _, dsn := range config.Destinations.Node {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case CSHARP:
			for _, dsn := range config.Destinations.Csharp {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case DART:
			for _, dsn := range config.Destinations.Dart {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case ELIXIR:
			for _, dsn := range config.Destinations.Elixir {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case PERL:
			for _, dsn := range config.Destinations.Perl {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case RUST:
			for _, dsn := range config.Destinations.Rust {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case COCOA:
			for _, dsn := range config.Destinations.Cocoa {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		case ANDROID:
			for _, dsn := range config.Destinations.Android {
				event.setDsn(dsn)
				request := NewRequest(event)
				request.send()
			}
		default:
			sentry.CaptureMessage("unsupported event platform: " + event.Platform)
			fmt.Printf("\nunrecognized Platform %v\n", event.Platform)
		}
	}
	fmt.Printf("\n> TOTAL sent: %v", counter)

	// does not Capture, not sure why
	sentry.CaptureMessage("finished sending all requests")
}
