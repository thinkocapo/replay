package main

import (
	"database/sql"
	"bytes"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"compress/gzip"
	"strconv"

	"github.com/buger/jsonparser"
	"io/ioutil"
	// "encoding/json"
	// "io/ioutil"
	// "log"
	// "net/http"
	// "os"
	// "time"
	// "github.com/getsentry/sentry-go"
	// sentryhttp "github.com/getsentry/sentry-go/http"
)

// type Event struct {
// 	id         int
// 	name   string
// 	// type    string
// 	payload []byte
// 	headers []byte
// }

// func getEvent(db *sql.DB, id2 int) User {
// 	rows, err := db.Query("select * from testTable")
// 	checkError(err)
// 	for rows.Next() {
// 	  var tempUser User
// 	  err =
// 		rows.Scan(&tempUser.id, &tempUser.username, &tempUser.surname, &tempUser.age, &tempUser.university)
// 	  checkError(err)
// 	  if tempUser.id == id2 {
// 		return tempUser
// 	  }
// 	}
// 	return User{}
// }

// function getEvent(sql.rows rows) int {
// 	for rows.Next() {
// 		// var event Event
// 		var id int
// 		// var name string
// 		// var body []bytes
// 		// var headers []bytes

// 		err =
// 			  rows.Scan(&id)
// 		// checkError(err)
// 		// if tempUser.id == id2 {
// 		// 	return tempUser
// 		// }
// 		return id
// 	}
// }

func main() {


	fmt.Println("Let's connect Sqlite")
	db, _ := sql.Open("sqlite3", "sqlite.db")
	fmt.Println("Let's connect Sqlite", db)

	rows, err := db.Query("SELECT * FROM events")
	if err != nil {
		fmt.Println("We got Error", err)
	}
	
	for rows.Next() {
		var id int
		var name string
		var typee string
		var body []byte
		var headers []byte
		rows.Scan(&id, &name, &typee, &body, &headers)
		
		r, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			fmt.Println(err)
		}
		body, err = ioutil.ReadAll(r)
		if err != nil {
			fmt.Println(err)
		}
		
		// Sentry Event - String Types
		event_id, err := jsonparser.GetString(body, "event_id")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("id", strconv.Itoa(id), event_id)

	}

	// fmt.Println("LENGTH", len(rows))
	rows.Close()
}
// https://dev.to/fevziomurtekin/using-sqlite-in-go-programming-3g2c
// https://www.thepolyglotdeveloper.com/2017/04/using-sqlite-database-golang-application/