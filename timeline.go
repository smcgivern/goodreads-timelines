package main

import (
	"fmt"
	"github.com/KyleBanks/goodreads"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/yosida95/uritemplate"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

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

func userShow(c *cache.Cache, client *goodreads.Client, userId string) (*goodreads.User, error) {
	key := fmt.Sprintf("UserShow:%s", userId)
	fromCache, found := c.Get(key)

	if found {
		userInfo, ok := fromCache.(*goodreads.User)

		// Cached value may be a pointer or not depending on
		// whether cache came from memory or file
		if ok {
			return userInfo, nil
		} else {
			tmp := fromCache.(goodreads.User)
			return &tmp, nil
		}
	} else {
		userInfo, err := client.UserShow(userId)
		if err != nil {
			return nil, err
		}

		c.Set(key, userInfo, cache.DefaultExpiration)

		return userInfo, nil
	}
}

func reviewPage(c *cache.Cache, client *goodreads.Client, userId string, page int) ([]goodreads.Review, error) {
	var err error
	key := fmt.Sprintf("ReviewList:%s:%s", userId, page)
	reviews, found := c.Get(key)

	if !found {
		reviews, err = client.ReviewList(userId, "read", "date_read", "", "a", page, 200)
		if err != nil {
			return nil, err
		}

		c.Set(key, reviews, cache.DefaultExpiration)
	}

	return reviews.([]goodreads.Review), nil
}

func timeline(c *cache.Cache, client *goodreads.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := mux.Vars(r)["userId"]
		userInfo, err := userShow(c, client, userId)

		if err != nil {
			http.Error(w, "Failed to load user info", http.StatusInternalServerError)
			return
		}

		reviews := make([]goodreads.Review, 0)
		template := template.Must(template.New("layout.html").Funcs(functionMap).ParseFiles("template/layout.html", "template/timeline.html"))
		vars := uritemplate.Values{}
		vars.Set("user_id", uritemplate.String(userId))
		vars.Set("user_name", uritemplate.String(strings.ToLower(userInfo.Name)))

		userLink, err := userLinkTemplate.Expand(vars)
		if err != nil {
			http.Error(w, "Failed to generate user link", http.StatusInternalServerError)
			return
		}

		for page := 1; page < 100; page++ {
			reviewPage, err := reviewPage(c, client, userId, page)
			if err != nil {
				http.Error(w, "Failed to get reviews", http.StatusInternalServerError)
				return
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
}
