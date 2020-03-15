
package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"encoding/json"
)

type Event struct {
	Platform string
	Level string
	Server_name string
}


func DecodeRequest(w http.ResponseWriter, req *http.Request) {

	// requestDump, err := httputil.DumpRequest(req, true)
	// fmt.Print(string(requestDump))
	// sends as the http response
	// fmt.Fprint(w, string(requestDump))
	// if err != nil {
	// 	fmt.Fprint(w, err.Error())
	// } else {
	// }

	var event Event
	err := json.NewDecoder(req.Body).Decode(&event)
	if err != nil {
		fmt.Printf("%v", "\n------- ERRROR --------\n")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		fmt.Printf("%v", "\n------- NO ERROR -------\n")
		fmt.Printf("%v", event)
	}

}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/decoderequest", DecodeRequest).Methods("GET")
	router.HandleFunc("/decoderequest", DecodeRequest).Methods("POST")
	log.Fatal(http.ListenAndServe(":12345", router))
}