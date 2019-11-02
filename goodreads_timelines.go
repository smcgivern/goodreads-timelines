package main

import (
	"bytes"
	"fmt"
	"github.com/KyleBanks/goodreads"
	"github.com/gorilla/mux"
	libsass "github.com/wellington/go-libsass"
	"github.com/yosida95/uritemplate"
	"golang.org/x/text/message"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
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
	ByMonth      [][][]goodreads.Review
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
		log.Fatal(err)
	}

	return t
}

// TODO: simplify! (Break into smaller functions? Define types for slices of reviews?)
func byMonth(startByMonth time.Time, finish time.Time, reviews []goodreads.Review) [][][]goodreads.Review {
	currentDate := startByMonth
	finishByMonth := time.Date(finish.Year(), finish.Month(), 1, 23, 59, 59, 0, time.UTC).AddDate(0, 1, -1)
	length := daysBetween(finishByMonth, currentDate) + 1
	reviewsByMonth := make([][][]goodreads.Review, 1)
	reviewsByMonth[0] = make([][]goodreads.Review, 1)
	reviewsByMonth[0][0] = []goodreads.Review{}
	currentReview := reviews[0]
	month := 0
	day := 0
	review := 0

	for {
		dateDiff := currentDate.AddDate(0, 0, 1).Sub(parseTime(currentReview.ReadAt)).Hours()

		if review < len(reviews) && dateDiff > 0 && dateDiff < 24 {
			dayIndex := currentDate.Day() - 1
			review++
			reviewsByMonth[month][dayIndex] = append(reviewsByMonth[month][dayIndex], currentReview)

			if review < len(reviews) {
				currentReview = reviews[review]
			}
		} else {
			day++
			currentDate = currentDate.AddDate(0, 0, 1)

			if day == length {
				break
			}

			if currentDate.Day() == 1 {
				month++
				reviewsByMonth = append(reviewsByMonth, [][]goodreads.Review{})
			}
			reviewsByMonth[month] = append(reviewsByMonth[month], []goodreads.Review{})
		}
	}

	return reviewsByMonth
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
	reviews := make([]goodreads.Review, 0)

	userInfo, err := client.UserShow(userId)
	if err != nil {
		log.Fatal(err)
	}

	template := template.Must(template.New("layout.html").Funcs(functionMap).ParseFiles("template/layout.html", "template/timeline.html"))
	vars := uritemplate.Values{}
	vars.Set("user_id", uritemplate.String(userId))
	vars.Set("user_name", uritemplate.String(strings.ToLower(userInfo.Name)))

	userLink, err := userLinkTemplate.Expand(vars)
	if err != nil {
		log.Fatal(err)
	}

	for page := 1; page < 100; page++ {
		reviewPage, err := client.ReviewList(userId, "read", "date_read", "", "a", page, 200)
		if err != nil {
			log.Fatal(err)
		}

		if len(reviewPage) == 0 {
			break
		}

		for i := range reviewPage {
			if reviewPage[i].ReadAt != "" {
				reviews = append(reviews, reviewPage[i])
			}
		}
	}

	reviewLength := len(reviews)
	start := parseTime(reviews[0].ReadAt)
	finish := parseTime(reviews[reviewLength-1].ReadAt)
	startByMonth := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)

	page := Page{
		Title: fmt.Sprintf("Goodreads timeline for %s", userInfo.Name),
		Scripts: []string{"/ext/jquery-1.7.min.js", "/ext/jquery-1.7.min.js", "/ext/flot.min.js",
			"/ext/qtip.min.js", "/ext/chart.js", "/ext/tooltip.js"},
		UserInfo:     *userInfo,
		UserLink:     userLink,
		Start:        start,
		Finish:       finish,
		ReviewLength: reviewLength,
		StartByMonth: startByMonth,
		ByMonth:      byMonth(startByMonth, finish, reviews),
	}

	err = template.Execute(w, page)
	if err != nil {
		log.Println(err)
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
		"baseUrl":     baseUrl,
		"thousands":   thousands,
		"parseTime":   parseTime,
		"daysBetween": daysBetween,
		"isoDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"perWeek": func(books int, days int) float64 {
			return float64(books*7) / float64(days)
		},
		"makeSlice": func(length int) []int {
			return make([]int, length)
		},
		"reviewLength": func(days [][]goodreads.Review) int {
			sum := 0
			for _, day := range days {
				sum += len(day)
			}
			return sum
		},
		"inc": func(x int) int {
			return x + 1
		},
		"offset": func(t time.Time) int {
			return int(t.Weekday())
		},
		"pointFor": func(week int, day int) int {
			return (week * 7) + day
		},
		"between": func(point int, offset int, first time.Time, last time.Time) bool {
			return (point+1) >= (offset+first.Day()) && (point+1) <= (offset+last.Day())
		},
		"dateForPoint": func(point int, offset int, first time.Time) time.Time {
			return first.AddDate(0, 0, (point - offset))
		},
		"reviewsFor": func(date time.Time, days [][]goodreads.Review) []goodreads.Review {
			return days[date.Day()-1]
		},
	}

	r := mux.NewRouter()

	r.HandleFunc(baseUrl("/"), home)
	r.HandleFunc(baseUrl("/go-to-timeline/"), goToTimeline)
	r.HandleFunc(baseUrl("/ext/style.css"), stylesheet(compileSass()))
	r.PathPrefix(baseUrl("/ext/")).Handler(http.StripPrefix(baseUrl("/ext/"), http.FileServer(http.Dir("public/ext"))))
	r.HandleFunc(baseUrl("/:{userId}/"), timeline)

	log.Fatal(http.ListenAndServe(":8080", r))
}
