package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

type DSN struct {
	host      string
	rawurl    string
	key       string
	projectId string
}

func NewDSN(rawurl string) *DSN {
	fmt.Println("> rawlurl", rawurl)

	// still need support for http vs. https 7: vs 8:
	key := strings.Split(rawurl, "@")[0][7:]

	uri, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	idx := strings.LastIndex(uri.Path, "/")
	if idx == -1 {
		log.Fatal("missing projectId in dsn")
	}
	projectId := uri.Path[idx+1:]

	var host string
	if strings.Contains(rawurl, "ingest.sentry.io") {
		// TODO slice the o87286 dynamically
		host = "o87286.ingest.sentry.io"
	}
	if strings.Contains(rawurl, "@localhost:") {
		host = "localhost:9000"
	}
	if host == "" {
		log.Fatal("missing host")
	}
	// if len(key) < 31 || len(key) > 32 {
	// 	log.Fatal("bad key length")
	// }
	if projectId == "" {
		log.Fatal("missing project Id")
	}
	fmt.Printf("> DSN { host: %s, projectId: %s }\n", host, projectId)
	return &DSN{
		host,
		rawurl,
		key,
		projectId,
	}
}

func (d DSN) storeEndpoint() string {
	var fullurl string
	if strings.Contains(d.host, "ingest.sentry.io") {
		// [1:] is for removing leading slash from sentry_key=/a971db611df44a6eaf8993d994db1996, which errors ""bad sentry DSN public key""
		fullurl = fmt.Sprint("https://", d.host, "/api/", d.projectId, "/store/?sentry_key=", d.key[1:], "&sentry_version=7")
		// fullurl = fmt.Sprint("https://", d.host, "/api/", d.projectId, "/store/?sentry_key=", d.key[1:])
	}
	if d.host == "localhost:9000" {
		fullurl = fmt.Sprint("http://", d.host, "/api/", d.projectId, "/store/?sentry_key=", d.key, "&sentry_version=7")
	}
	if fullurl == "" {
		log.Fatal("problem with fullurl")
	}
	return fullurl
}

func (d DSN) envelopeEndpoint() string {
	var fullurl string
	if strings.Contains(d.host, "ingest.sentry.io") {
		fullurl = fmt.Sprint("https://", d.host, "/api/", d.projectId, "/envelope/?sentry_key=", d.key[1:], "&sentry_version=7")
	}
	if d.host == "localhost:9000" {
		fullurl = fmt.Sprint("http://", d.host, "/api/", d.projectId, "/envelope/?sentry_key=", d.key, "&sentry_version=7")
	}
	if fullurl == "" {
		log.Fatal("problem with fullurl")
	}
	return fullurl
}
