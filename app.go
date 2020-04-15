package main
// This is for sending an Event directly to your Sentry instance. does not use proxy. tesing purposes.
import (
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	"os"
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	// sentryhttp "github.com/getsentry/sentry-go/http"
)


// TODO - generate Go exceptions and capture via sentry sdk

// use cli args for # of errors sent. cap it at 100
func main() {

	// _ = sentry.Init(sentry.ClientOptions{
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "http://42e526238f594d69888875c6f10e261f@localhost:9000/3",
		// Release:          os.Args[1], Release based on day that job is running.
		// AttachStacktrace: true,
		// ServerName:       "SE1.US.EAST"
		//Debug:       false,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	defer sentry.Flush(2 * time.Second)

	resp, err := http.Get(os.Args[1])
	if err != nil {
		sentry.CaptureException(err)
		log.Printf("reported to Sentry: %s", err)
		return
	}
	defer resp.Body.Close()

	// sentryHandler := sentryhttp.New(sentryhttp.Options{
	// 	Repanic: true,
	// })

	// Flush buffered events before the program terminates.
	// defer sentry.Flush(2 * time.Second)
	// sentry.CaptureMessage("It works!")

	fmt.Println("Hello, 世界")

}
