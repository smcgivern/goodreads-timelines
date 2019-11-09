package main

import (
	"fmt"
	"golang.org/x/text/message"
	"io/ioutil"
	"strings"
	"time"
)

func readFile(path string) string {
	file, err := ioutil.ReadFile(path)
	if err == nil {
		return strings.TrimSpace(string(file))
	} else {
		return ""
	}
}

func baseUrl(url string) string {
	rootLeadingSlash := fmt.Sprintf("/%s", rootUrl)
	rootDoubleSlash := fmt.Sprintf("%s/", rootLeadingSlash)

	if strings.HasPrefix(url, "/") && !strings.HasPrefix(url, rootDoubleSlash) {
		return fmt.Sprintf("%s%s", rootLeadingSlash, url)
	} else {
		return url
	}
}

func thousands(n int) string {
	p := message.NewPrinter(message.MatchLanguage("en"))
	return p.Sprint(n)
}

func daysBetween(a time.Time, b time.Time) int {
	return int(a.Sub(b).Hours() / 24)
}

func parseTime(d string) time.Time {
	t, err := time.Parse(time.RubyDate, d)
	if err != nil {
		bail(err)
	}

	return t
}
