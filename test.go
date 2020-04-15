package main

import (
	"database/sql"
	// "encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	// "io/ioutil"
	// "os"
	// "log"
	// "net/http"
	// "time"
	// "github.com/getsentry/sentry-go"
	// sentryhttp "github.com/getsentry/sentry-go/http"
)


// TODO - generate Go exceptions and capture via sentry sdk

// use cli args for # of errors sent. cap it at 100
func main() {


	fmt.Println("Let's connect Sqlite")
	db, _ := sql.Open("sqlite3", "sqlite.db")

	fmt.Println("Let's connect Sqlite", db)

	rows, err := db.Query("select * from events")
	if err != nil {
		fmt.Println("We got Error", err)
	}
	fmt.Println("LENGTH", rows)

}
