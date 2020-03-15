/*
This middleware made for auth system that randomly generate access tokens, which used later for accessing secure content. Since there is no pre-defined token value, naive approach without middleware (or if middleware use only request payloads) will fail, because replayed server have own tokens, not synced with origin. To fix this, our middleware should take in account responses of replayed and origin server, store `originalToken -> replayedToken` aliases and rewrite all requests using this token to use replayed alias. See `middleware_test.go#TestTokenMiddleware` test for examples of using this middleware.
How middleware works:
                   Original request      +--------------+
+-------------+----------STDIN---------->+              |
|  Gor input  |                          |  Middleware  |
+-------------+----------STDIN---------->+              |
                   Original response     +------+---+---+
                                                |   ^
+-------------+    Modified request             v   |
| Gor output  +<---------STDOUT-----------------+   |
+-----+-------+                                     |
      |                                             |
      |            Replayed response                |
      +------------------STDIN----------------->----+
*/

package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/base64"
	"fmt"
	"os"
	"compress/gzip"
	"compress/zlib"
	"github.com/buger/goreplay/proto"
	"github.com/buger/jsonparser"
	"io/ioutil"
	"encoding/csv"
	"log"
)

var HTTP_CONTENT_ENCODING = []byte("Content-Encoding")
var ENCODING_GZIP = []byte("gzip")
var ENCODING_DEFLATE = []byte("deflate")


// requestID -> originalToken
var originalTokens map[string][]byte

// originalToken -> replayedToken
var tokenAliases map[string][]byte

func main() {
	originalTokens = make(map[string][]byte)
	tokenAliases = make(map[string][]byte)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		encoded := scanner.Bytes()
		buf := make([]byte, len(encoded)/2)
		hex.Decode(buf, encoded)

		process(buf)
	}
}

type Event struct {
	Platform string
	Level string
	Event_id string
	Timestamp string
	Server_name string
}

func process(buf []byte) (error) {
	// Headers - User-Agent, HostAddress, ClientIP, ClientPort, HTTPProtocalVersion, Connection 
	// Body - all, is from sentry_sdk
	headerSize := bytes.IndexByte(buf, '\n') + 1
	payload := buf[headerSize:]
	// Debug("Received payload:", string(buf))
	end := proto.MIMEHeadersEndPos(payload)
	body := payload[end:]	

	// doesn't work on the 'body' object
	// var event Event
	// err := json.NewDecoder(body).Decode(&event)
	// if err != nil {
	// 	fmt.Printf("%v", "\n------- ERRROR --------\n")
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// } else {
	// 	fmt.Printf("%v", event) // logs Platform, Level, Server_name as platform, level, server_name
	// }

	body, err := decodeBody(body, proto.Header(payload, HTTP_CONTENT_ENCODING))
	if err != nil {
		return err
	}

	// Sentry Event - String Types
	platform, err := jsonparser.GetString(body, "platform")
	level, err := jsonparser.GetString(body, "level")
	event_id, err := jsonparser.GetString(body, "event_id")
	timestamp, err := jsonparser.GetString(body, "timestamp")
	server_name, err := jsonparser.GetString(body, "server_name")

	// Sentry Event - Other Types
	// could save a stringfied object/array to start

	var event = Event{platform, level, event_id, timestamp, server_name} 
	Debug("event", event)

	// Persist to CSV/DB
	records := [][]string{
		{"event_id", "server_name", "platform", "level", "timestamp"},
		{event_id, server_name, platform, level, timestamp},
		{"write", "these", "in", "batches", "maybe"},
	}

	file, err := os.Create("result.csv")
	w := csv.NewWriter(file)
	w.WriteAll(records) // calls Flush internally
	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}

	return nil
}

func encode(buf []byte) []byte {
	dst := make([]byte, len(buf)*2+1)
	hex.Encode(dst, buf)
	dst[len(dst)-1] = '\n'

	return dst
}

func Debug(args ...interface{}) {
	fmt.Fprint(os.Stderr, "[DEBUG][TOKEN-MOD] ")
	fmt.Fprintln(os.Stderr, args...)
}

func decodeBody(body []byte, contentEncoding []byte) ([]byte, error) {
	switch {
	case bytes.Equal(contentEncoding, ENCODING_GZIP):
		fmt.Printf("%v", "11111111111")
		r, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		body, err = ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return body, nil
	case bytes.Equal(contentEncoding, ENCODING_DEFLATE):
		fmt.Printf("%v", "2222222222")
		r, err := zlib.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		body, err = ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return body, nil
	case len(body) > 0 && body[0] != '{':
		fmt.Printf("%v", "3333333333")
		buf := make([]byte, base64.StdEncoding.DecodedLen(len(body)))
		n, err := base64.StdEncoding.Decode(buf, body)
		if err != nil {
			return nil, err
		}
		body = buf[:n]
		r, err := zlib.NewReader(bytes.NewReader(body))
		if err != nil {
			return body, nil
		}
		buf, err = ioutil.ReadAll(r)
		if err != nil {
			return body, nil
		}
		return buf, nil
	}
	fmt.Printf("%v", "444444444")
	return body, nil
}

type BodyEncoder func([]byte) []byte

func gzipEncoder(b []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(b)
	w.Close()
	return buf.Bytes()
}

func b64DeflateEncoder(b []byte) []byte {
	return b64Encoder(deflateEncoder(b))
}

func b64Encoder(b []byte) []byte {
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(buf, b)
	return buf
}

func noopEncoder(b []byte) []byte { return b }

func deflateEncoder(b []byte) []byte {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(b)
	w.Close()
	return buf.Bytes()
}