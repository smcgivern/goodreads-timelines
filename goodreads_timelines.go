package main

import (
	"fmt"
	libsass "github.com/wellington/go-libsass"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Page struct {
	Title   string
	Scripts []string
}

var rootUrl string
var functionMap template.FuncMap

func baseUrl(url string) string {
	rootLeadingSlash := fmt.Sprintf("/%s", rootUrl)
	rootDoubleSlash := fmt.Sprintf("%s/", rootLeadingSlash)

	if strings.HasPrefix(url, "/") && !strings.HasPrefix(url, rootDoubleSlash) {
		return fmt.Sprintf("%s%s", rootLeadingSlash, url)
	} else {
		return url
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	template := template.Must(template.New("layout.html").Funcs(functionMap).ParseFiles("template/layout.html", "template/index.html"))

	page := Page{
		Title:   "Goodreads timelines",
		Scripts: []string{"/ext/jquery-1.7.min.js", "/ext/index.js"},
	}

	template.Execute(w, page)
}

func stylesheet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")

	file, err := os.Open("template/css/style.scss")
	if err != nil {
		log.Fatal(err)
	}

	compiler, err := libsass.New(w, file)
	if err != nil {
		log.Fatal(err)
	}

	includePaths := []string{"template/css/"}
	err = compiler.Option(libsass.IncludePaths(includePaths))
	if err != nil {
		log.Fatal(err)
	}

	if err := compiler.Run(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	rootFile, err := ioutil.ReadFile(".root")
	if err != nil {
		rootUrl = ""
	} else {
		rootUrl = strings.TrimSpace(string(rootFile))
	}

	functionMap = template.FuncMap{
		"baseUrl": baseUrl,
	}

	http.HandleFunc(baseUrl("/"), homePage)
	http.Handle(baseUrl("/ext/"),
		http.StripPrefix(baseUrl("/ext/"), http.FileServer(http.Dir("public/ext"))))
	http.HandleFunc(baseUrl("/ext/style.css"), stylesheet)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
