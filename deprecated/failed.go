	// attempt 2 - failed
	// bodyBuffer, _ := ioutil.ReadAll(req.Body)
	// // Put the body back for FormatRequest to read it // io/ioutil
	// req.Body = myReader{bytes.NewBuffer(buf)}
	// fmt.Printf("--> %s\n\n", formatRequest(req))
	
	// attempt 3 - failed
	// decoder := json.NewDecoder(req.Body)
    // var t Event
    // err := decoder.Decode(&t)
    // if err != nil {
	// 	log.Println(t.platform)
	// 	log.Println("\nBAAADDDDDD")
	// }
	

	// deprecated because switched to json:
		// attempt 1
	// req.ParseForm()
	// for key, value := range req.Form {
	// 	fmt.Printf("%s = %s\n", key, value)
	// }
	// run against a struct to type check it here
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


	// http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	// })
	// log.Fatal(http.ListenAndServe(":8080", nil))


	// buf, err := ioutil.ReadAll(req.Body) // io/ioutil
