package main

import (
	"bytes"
	"context"
	
	// "database/sql"
	_ "github.com/mattn/go-sqlite3"

	"fmt"
	"io"
	"io/ioutil"
	"time"

	"encoding/json"

	"cloud.google.com/go/storage"

	// "os"
	"strings"
)

func main () {
	fmt.Print("000000000000")

	bucket := "undertakerevents"
	object := "tracing-example-multiproject.db"

	var buf1 bytes.Buffer
	w := io.Writer(&buf1)

	downloadFile(w, bucket, object)
}


func downloadFile(w io.Writer, bucket, object string) ([]byte, error) {
	fmt.Println("\ndownloadFile")
	
	
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
			fmt.Println("ERROR", err)
			return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
			return nil, fmt.Errorf("Object(%q).NewReader: %v", object, err)
	}

	data, err := ioutil.ReadAll(rc)
	fmt.Print(string(data)) // can see text now


	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll: %v", err)
	}

	// myarray := strings.Split(string(data), "{\"") // %, \", \n
	// fmt.Println("lENGHT....", len(myarray))

	// if _, err := io.Copy(os.Stdout, rc); err != nil {
		// TODO: Handle error.
		// fmt.Print(err)
	// }
	// Prints "This object contains text."


	// fmt.Fprintf(w, "Blob %v downloaded.\n", object)
	// NOOOOO, does not work here. no files in Cloud Functions
	// database, _ = sql.Open("sqlite3", data)
	// database.Query()

	// PLAN OF ATTACK
	// Can open sqlite by passing the response object rathen than sql.Open() a a flatfile.db?
	// What else can do with rc instead of passing it to ioutil.ReadAll(rc)?
	// Is a 'blob' how to work with blob? from ioutil.ReadAll(rc)
		// convert to .newFile?
		// make a file object using the rc/data?
	// gobyexample

	// DownloadFile | Getting/Reading it in a different manner?
	
	// Readings FlatFile in Go / Reading FlatFiles in a CLoud Function
	// Reaing FlatFile Data from Blob
	// CloudStorage'ing a flat-file and parsing it, deserialize?

	// What does Sql.Open return?
	
	// AppEngine can save/write the file?

	return data, nil
}

func unmarshalJSON(bytes []byte) map[string]interface{} {
	var _interface map[string]interface{}
	if err := json.Unmarshal(bytes, &_interface); err != nil {
		panic(err)
	}
	return _interface
}
