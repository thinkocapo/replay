
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
    platform []string
}
// type Person struct {
// 	name string
// 	age string
// }


func DumpRequest(w http.ResponseWriter, req *http.Request) {
	// attempt 1
	req.ParseForm()
	for key, value := range req.Form {
		fmt.Printf("%s = %s\n", key, value)
	}
	// TODO run against a struct to type check it here
	// Looks like all are arrays
	/*
		$ ./dump-request 
		exception = [values]
		event_id = [6f6755c11b764cbcb523ba165b062c4c]
		timestamp = [2020-03-15T00:13:47.319687Z]
		extra = [sys.argv]
		contexts = [runtime]
		platform = [python]
		level = [error]
		modules = [urllib3 sentry-sdk requests python-dotenv idna chardet certifi xkit wheel virtualenv ufw ubuntu-drivers-common systemd-python system76driver system-service six setuptools sessioninstaller secretstorage screen-resolution-extra requests-unixsocket reportlab repoman pyyaml pyxdg pytz python-xlib python-debian python-apt pyrfc3339 pynacl pymacaroons pygobject pydbus pycups pycrypto pycairo psutil protobuf powerline-status pip pillow pexpect olefile netifaces macaroonbakery louis language-selector keyrings.alt keyring kernelstub httplib2 hidpidaemon evdev distro-info defer cupshelpers cryptography command-not-found chrome-gnome-shell brlapi asn1crypto]
		server_name = [pop-os]
		sdk = [name version packages integrations]
	*/


	// TODO - deprecate this since DumpRequest doesn't show the x-www-form-urlencoded data. It will never be on req.Body below:
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Fprint(w, err.Error())
	} else {
		var event Event
		err := json.NewDecoder(req.Body).Decode(&event)
		if err != nil {
			fmt.Printf("%v", "\nERRROR\n")
			http.Error(w, err.Error(), http.StatusBadRequest)
			// return
		} else {
			fmt.Printf("%v", "NO ERROR")
			// sends as the http response
			// fmt.Fprint(w, string(requestDump))
		}
	
	}

	fmt.Print(string(requestDump))
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/dumprequest", DumpRequest).Methods("GET")
	router.HandleFunc("/dumprequest", DumpRequest).Methods("POST")
	log.Fatal(http.ListenAndServe(":12345", router))



	// http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	// })
	// log.Fatal(http.ListenAndServe(":8080", nil))
}