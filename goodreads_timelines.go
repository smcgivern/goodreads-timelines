package main

import (
	"bytes"
	"fmt"
	"github.com/KyleBanks/goodreads"
	"github.com/gorilla/mux"
	libsass "github.com/wellington/go-libsass"
	"github.com/yosida95/uritemplate"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Page struct {
	Title    string
	Scripts  []string
	UserId   string
	UserLink string
}

var rootUrl string
var goodreadsKey string
var userLinkTemplate *uritemplate.Template
var functionMap template.FuncMap

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

func compileSass() *bytes.Reader {
	var buffer bytes.Buffer

	file, err := os.Open("template/css/style.scss")
	if err != nil {
		log.Fatal(err)
	}

	compiler, err := libsass.New(&buffer, file)
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

	return bytes.NewReader(buffer.Bytes())
}

func timeline(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["userId"]
	client := goodreads.NewClient(goodreadsKey)

	userInfo, err := client.UserShow(userId)
	if err != nil {
		log.Fatal(err)
	}

	template := template.Must(template.New("layout.html").Funcs(functionMap).ParseFiles("template/layout.html", "template/timeline.html"))
	vars := uritemplate.Values{}
	vars.Set("user_id", uritemplate.String(userId))

	userLink, err := userLinkTemplate.Expand(vars)
	if err != nil {
		log.Fatal(err)
	}

	page := Page{
		Title: fmt.Sprintf("Goodreads timeline for %s", userInfo.Name),
		Scripts: []string{"/ext/jquery-1.7.min.js", "/ext/jquery-1.7.min.js", "/ext/flot.min.js",
			"/ext/qtip.min.js", "/ext/chart.js", "/ext/tooltip.js"},
		UserId:   userId,
		UserLink: userLink,
	}

	template.Execute(w, page)
}

func goToTimeline(w http.ResponseWriter, r *http.Request) {
	userId := userLinkTemplate.Match(r.FormValue("goodreads-uri")).Get("user_id")
	http.Redirect(w, r, baseUrl(fmt.Sprintf("/:%s/", userId)), http.StatusSeeOther)
}

func home(w http.ResponseWriter, r *http.Request) {
	template := template.Must(template.New("layout.html").Funcs(functionMap).ParseFiles("template/layout.html", "template/index.html"))

	page := Page{
		Title:   "Goodreads timelines",
		Scripts: []string{"/ext/jquery-1.7.min.js", "/ext/index.js"},
	}

	template.Execute(w, page)
}

func stylesheet(buffer *bytes.Reader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		buffer.WriteTo(w)
		buffer.Seek(0, 0)
	}
}

func main() {
	rootUrl = readFile(".root")
	goodreadsKey = readFile("goodreads.key")
	userLinkTemplate = uritemplate.MustNew("https://www.goodreads.com/user/show/{user_id}-{user_name}")
	functionMap = template.FuncMap{
		"baseUrl": baseUrl,
	}

	r := mux.NewRouter()

	r.HandleFunc(baseUrl("/"), home)
	r.HandleFunc(baseUrl("/go-to-timeline/"), goToTimeline)
	r.HandleFunc(baseUrl("/ext/style.css"), stylesheet(compileSass()))
	r.PathPrefix(baseUrl("/ext/")).Handler(http.StripPrefix(baseUrl("/ext/"), http.FileServer(http.Dir("public/ext"))))
	r.HandleFunc(baseUrl("/:{userId}/"), timeline)

	log.Fatal(http.ListenAndServe(":8080", r))
}
