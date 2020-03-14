
package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httputil"
	"encoding/json"
)
type Person struct {
    Name string
    Age  int
}

func DumpRequest(w http.ResponseWriter, req *http.Request) {
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Fprint(w, err.Error())
	} else {
		// Declare a new Person struct.
		var p Person

		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(req.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	
		// Do something with the Person struct...
		fmt.Printf("%v", "hey")
		fmt.Fprintf(w, "Person: %+v", p)
		fmt.Fprint(w, string(requestDump))
	}
}

func main() {
	router := mux.NewRouter()
	// sentry_sdk's before_send callback redirects the sdk's outbound traffic to /dumprequest
	router.HandleFunc("/dumprequest", DumpRequest).Methods("GET")
	router.HandleFunc("/dumprequest", DumpRequest).Methods("POST")

	log.Fatal(http.ListenAndServe(":12345", router))
}