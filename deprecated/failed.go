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