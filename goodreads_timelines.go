package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	libsass "github.com/wellington/go-libsass"
)

type Page struct {
	Title   string
	Scripts []string
}

func homePage(w http.ResponseWriter, r *http.Request) {
	template := template.Must(template.ParseFiles("template/layout.html",
		"template/index.html"))

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
	http.HandleFunc("/", homePage)
	http.Handle("/ext/", http.StripPrefix("/ext/", http.FileServer(http.Dir("public/ext"))))
	http.HandleFunc("/ext/style.css", stylesheet)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
