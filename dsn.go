package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/getsentry/sentry-go"
)

type DSN struct {
	host      string
	rawurl    string
	key       string
	projectId string
}

func NewDSN(rawurl string) *DSN {
	// still need support for http vs. https 7: vs 8:
	key := strings.Split(rawurl, "@")[0][7:]

	uri, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	idx := strings.LastIndex(uri.Path, "/")
	if idx == -1 {
		sentry.CaptureMessage("missing projectId in dsn")
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
		sentry.CaptureMessage("missing host")
		log.Fatal("missing host")
	}
	// if len(key) < 31 || len(key) > 32 {
	// 	log.Fatal("bad key length")
	// }
	if len(projectId) != 7 {
		sentry.CaptureException(errors.New("bad project Id in dsn" + projectId))
		log.Fatal("bad project Id in dsn")
	}
	if projectId == "" {
		sentry.CaptureMessage("missing project Id")
		log.Fatal("missing project Id")
	}
	return &DSN{
		host,
		rawurl,
		key,
		projectId,
	}
}

func (d DSN) storeEndpoint() string {
	// [1:] is for removing leading slash from sentry_key=/a971db611df44a6eaf8993d994db1996
	fullurl := fmt.Sprintf("https://%v/api/%v/store/?sentry_key=%v&sentry_version=7", d.host, d.projectId, d.key[1:])

	if d.host == "localhost:9000" {
		fullurl = strings.Replace(fullurl, "http", "https", 1)
	}
	if fullurl == "" {
		sentry.CaptureMessage("missing fullurl")
		log.Fatal("missing fullurl")
	}
	return fullurl
}
