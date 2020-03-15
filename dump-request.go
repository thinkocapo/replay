
package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httputil"
	"encoding/json"
)

type Event struct {
	platform string
	shouldfail []string
}


func DumpRequest(w http.ResponseWriter, req *http.Request) {

	// TODO - deprecate this since DumpRequest doesn't show the x-www-form-urlencoded data. It will never be on req.Body below:
	requestDump, err := httputil.DumpRequest(req, true)
	fmt.Print(string(requestDump))
	// sends as the http response
	// fmt.Fprint(w, string(requestDump))


	if err != nil {
		fmt.Fprint(w, err.Error())
	} else {
		var event Event
		err := json.NewDecoder(req.Body).Decode(&event)
		if err != nil {
			fmt.Printf("%v", "\n------- ERRROR --------\n")
			http.Error(w, err.Error(), http.StatusBadRequest)
			// return
		} else {
			fmt.Printf("%v", "------- NO ERROR -------\n")
		}
	}

}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/dumprequest", DumpRequest).Methods("GET")
	router.HandleFunc("/dumprequest", DumpRequest).Methods("POST")
	log.Fatal(http.ListenAndServe(":12345", router))
}