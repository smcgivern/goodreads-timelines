package main

import (
	"encoding/gob"
	"fmt"
	"github.com/KyleBanks/goodreads"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/yosida95/uritemplate"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Page struct {
	Title        string
	Scripts      []string
	UserLink     string
	UserInfo     goodreads.User
	Start        time.Time
	Finish       time.Time
	ReviewLength int
	StartByMonth time.Time
	ReviewsMap   map[string][]goodreads.Review
	Calendar     [][][]time.Time
}

var c *cache.Cache
var rootUrl string
var goodreadsKey string
var userLinkTemplate *uritemplate.Template
var functionMap template.FuncMap

func assignFunctionMap() {
	functionMap = template.FuncMap{
		"baseUrl":     baseUrl,
		"thousands":   thousands,
		"parseTime":   parseTime,
		"daysBetween": daysBetween,
		"isoDate":     isoDate,
		"perWeek": func(books int, days int) float64 {
			return float64(books*7) / float64(days)
		},
		"countReviews": func(weeks [][]time.Time, reviewsMap map[string][]goodreads.Review) int {
			sum := 0
			for _, week := range weeks {
				for _, day := range week {
					if !day.IsZero() {
						sum += len(reviewsMap[isoDate(day)])
					}
				}
			}

			return sum
		},
		"inc": func(x int) int {
			return x + 1
		},
	}
}

func goToTimeline(w http.ResponseWriter, r *http.Request) {
	userId := userLinkTemplate.Match(r.FormValue("goodreads-uri")).Get("user_id")
	http.Redirect(w, r, baseUrl(fmt.Sprintf("/:%s/", userId)), http.StatusSeeOther)
}

func home(w http.ResponseWriter, r *http.Request) {
	template := template.Must(template.New("layout.html").Funcs(functionMap).ParseFiles("template/layout.html", "template/index.html"))

	page := Page{
		Title:   "Goodreads timelines",
		Scripts: []string{"/ext/index.js"},
	}

	template.Execute(w, page)
}

func bail(e error) {
	cleanup()
	log.Fatal(e)
}

// Need to only register gob types once. Taken from
// https://github.com/patrickmn/go-cache/pull/16. As the mutex isn't
// public, this can't lock the cache :-(
func saveFile(fname string) error {
	fp, err := os.Create(fname)
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(fp)
	err = enc.Encode(c.Items())
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

func cleanup() {
	err := saveFile("goodreads.cache")
	if err != nil {
		log.Println(err)
	}
}

func logRequests(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	c = cache.New(24*time.Hour, 10*time.Minute)
	rootUrl = readFile(".root")
	goodreadsKey = readFile("goodreads.key")
	userLinkTemplate = uritemplate.MustNew("https://www.goodreads.com/user/show/{user_id}-{user_name}")
	assignFunctionMap()

	client := goodreads.NewClient(goodreadsKey)
	r := mux.NewRouter()
	channel := make(chan os.Signal)
	port, exists := os.LookupEnv("PORT")

	if !exists {
		port = "8080"
	}

	gob.Register([]goodreads.Review{})
	gob.Register(goodreads.User{})

	err := c.LoadFile("goodreads.cache")
	if err != nil {
		log.Println(err)
	}

	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-channel
		cleanup()
		os.Exit(1)
	}()

	r.HandleFunc(baseUrl("/"), home)
	r.HandleFunc(baseUrl("/go-to-timeline/"), goToTimeline)
	r.PathPrefix(baseUrl("/ext/")).Handler(http.StripPrefix(baseUrl("/ext/"), http.FileServer(http.Dir("public/ext"))))
	r.HandleFunc(baseUrl("/:{userId}/"), timeline(c, client))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), logRequests(r)))
}
