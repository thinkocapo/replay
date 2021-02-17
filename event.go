package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

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
	// may not need
	if event.Kind == "" {
		fmt.Println("> event.Kind", event.Kind)
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

func (event *Event) getPlatform() string {
	var platform string
	if event.Kind == TRANSACTION {
		platform = event.Transaction.Platform
	}
	if event.Kind == ERROR {
		platform = event.Error.Platform
	}
	if event.Kind == DEFAULT {
		platform = event.Error.Platform
	}
	if platform == "" {
		sentry.CaptureException(errors.New("no event platform set"))
		log.Fatalf("no event platform set")
	}
	return platform
}

func (event *Event) setDsn(dsn string) {
	event.DSN = NewDSN(dsn)
	if event.DSN == nil {
		sentry.CaptureException(errors.New("null DSN"))
		log.Fatal("null DSN")
	}
}

func (event *Event) setDsnGCS() {
	for _, platform := range platforms {
		if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == platform {
			event.Platform = platform
			break
		} else if event.Kind == TRANSACTION && event.Transaction.Platform == platform {
			event.Platform = platform
			break
		}
	}
	if event.Platform == "" {
		sentry.CaptureException(errors.New("event.Kind and Type condition not found" + event.Kind))
		log.Fatalf("setDsnGCS() event Kind: %v and Platform: %v not recognized", event.Kind, event.Platform)
	}

	// if event.Kind == TRANSACTION && event.Transaction.Platform == JAVASCRIPT {
	// 	event.Platform = JAVASCRIPT
	// } else if event.Kind == TRANSACTION && event.Transaction.Platform == PYTHON {
	// 	event.Platform = PYTHON
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == JAVASCRIPT {
	// 	event.Platform = JAVASCRIPT
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == PYTHON {
	// 	event.Platform = PYTHON
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == JAVA {
	// 	event.Platform = JAVA
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == RUBY {
	// 	event.Platform = RUBY
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == GO {
	// 	event.Platform = GO
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == PHP {
	// 	event.Platform = PHP
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == NODE {
	// 	event.Platform = NODE
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == CSHARP {
	// 	event.Platform = CSHARP
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == DART {
	// 	event.Platform = DART
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == ELIXIR {
	// 	event.Platform = ELIXIR
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == PERL {
	// 	event.Platform = PERL
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == RUST {
	// 	event.Platform = RUST
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == COCOA {
	// 	event.Platform = COCOA
	// } else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == ANDROID {
	// 	event.Platform = ANDROID
	// } else {
	// 	sentry.CaptureException(errors.New("event.Kind and Type condition not found" + event.Kind))
	// 	log.Fatalf("setDsnGCS() event Kind: %v and Platform: %v not recognized", event.Kind, event.Platform)
	// }
}

// TODO Do Not Repeat Yourself DRY
func (event *Event) setPlatform() {
	if event.Kind == TRANSACTION && event.Transaction.Platform == JAVASCRIPT {
		event.Platform = JAVASCRIPT
	} else if event.Kind == TRANSACTION && event.Transaction.Platform == PYTHON {
		event.Platform = PYTHON
	} else if event.Kind == TRANSACTION && event.Transaction.Platform == JAVA {
		event.Platform = JAVA
	} else if event.Kind == TRANSACTION && event.Transaction.Platform == RUBY {
		event.Platform = RUBY
	} else if event.Kind == TRANSACTION && event.Transaction.Platform == GO {
		event.Platform = GO
	} else if event.Kind == TRANSACTION && event.Transaction.Platform == PHP {
		event.Platform = PHP
	} else if event.Kind == TRANSACTION && event.Transaction.Platform == NODE {
		event.Platform = NODE
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == JAVASCRIPT {
		event.Platform = JAVASCRIPT
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == PYTHON {
		event.Platform = PYTHON
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == JAVA {
		event.Platform = JAVA
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == RUBY {
		event.Platform = RUBY
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == GO {
		event.Platform = GO
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == PHP {
		event.Platform = PHP
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == NODE {
		event.Platform = NODE
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == CSHARP {
		event.Platform = CSHARP
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == DART {
		event.Platform = DART
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == ELIXIR {
		event.Platform = ELIXIR
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == PERL {
		event.Platform = PERL
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == RUST {
		event.Platform = RUST
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == COCOA {
		event.Platform = COCOA
	} else if (event.Kind == ERROR || event.Kind == DEFAULT) && event.Error.Platform == ANDROID {
		event.Platform = ANDROID
	} else {
		sentry.CaptureException(errors.New("event.Kind and Type condition not found" + event.Kind))
		log.Fatal("setPlatform() event.Kind and type not recognized " + event.Kind + " " + event.Error.Platform)
	}
}

func (e Event) undertake() {
	if e.Kind == ERROR || e.Kind == DEFAULT {
		if e.Error.Tags == nil {
			e.Error.Tags = make([][]string, 0)
		}
		// EVAL if []Tags already has 'replay', then it gets duplicated in the array, but doesn't error in Sentry
		tagItem := []string{"replay", "replay"}
		e.Error.Tags = append(e.Error.Tags, tagItem)
	}
	if e.Kind == TRANSACTION {
		if e.Transaction.Tags == nil {
			e.Transaction.Tags = make([][]string, 0)
		}
		tagItem := []string{"replay", "replay"}
		e.Transaction.Tags = append(e.Transaction.Tags, tagItem)
	}
}
