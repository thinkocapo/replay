
// package main

// import (
// 	"fmt"
// 	"github.com/gorilla/mux"
// 	"log"
// 	"net/http"
// 	"encoding/json"
// )

// type Event struct {
// 	Platform string
// 	Level string
// 	Server_name string
// }

// func DecodeRequest(w http.ResponseWriter, req *http.Request) {

// 	var event Event
// 	err := json.NewDecoder(req.Body).Decode(&event)
// 	if err != nil {
// 		fmt.Printf("%v", "\n------- ERRROR --------\n")
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	} else {
// 		fmt.Printf("%v", event) // logs Platform, Level, Server_name as platform, level, server_name
// 	}

// }

// func main() {
// 	router := mux.NewRouter()
// 	router.HandleFunc("/decoderequest", DecodeRequest).Methods("GET")
// 	router.HandleFunc("/decoderequest", DecodeRequest).Methods("POST")
// 	log.Fatal(http.ListenAndServe(":12345", router))
// }