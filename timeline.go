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

func reviewsByDay(reviews []goodreads.Review) map[string][]goodreads.Review {
	byDay := make(map[string][]goodreads.Review)

	for i := range reviews {
		review := reviews[i]
		key := isoDate(parseTime(review.ReadAt))
		_, ok := byDay[key]
		if !ok {
			byDay[key] = make([]goodreads.Review, 0)
		}

		byDay[key] = append(byDay[key], review)
	}

	return byDay
}

func calendar(startByMonth time.Time, finish time.Time) [][][]time.Time {
	finishByMonth := time.Date(finish.Year(), finish.Month(), 1, 23, 59, 59, 0, time.UTC).AddDate(0, 1, -1)
	currentDate := startByMonth
	length := daysBetween(finishByMonth, currentDate) + 1
	m := -1
	w := 0
	months := make([][][]time.Time, 0)

	for i := 0; i < length; i++ {
		weekday := int(currentDate.Weekday())
		if currentDate.Day() == 1 {
			m++
			w = 0
			months = append(months, make([][][]time.Time, 1)...)
			months[m] = append(months[m], make([][]time.Time, 1)...)
			months[m][w] = make([]time.Time, 7)
		} else if weekday == 0 {
			w++
			months[m] = append(months[m], make([][]time.Time, 1)...)
			months[m][w] = make([]time.Time, 7)
		}

		months[m][w][weekday] = currentDate
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return months
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
				review := reviewPage[i]
				if review.ReadAt != "" && review.ReadCount > 0 {
					reviews = append(reviews, review)
				}
			}
		}

		reviewLength := len(reviews)
		start := parseTime(reviews[0].ReadAt)
		finish := parseTime(reviews[reviewLength-1].ReadAt)
		startByMonth := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)

		page := Page{
			Title: fmt.Sprintf("Goodreads timeline for %s", userInfo.Name),
			Scripts: []string{
				"/ext/moment.min.js", "/ext/Chart.min.js",
				"/ext/popper.min.js", "/ext/tippy.umd.min.js",
				"/ext/chart.js",
			},
			UserInfo:     *userInfo,
			UserLink:     userLink,
			Start:        start,
			Finish:       finish,
			ReviewLength: reviewLength,
			StartByMonth: startByMonth,
			ReviewsMap:   reviewsByDay(reviews),
			Calendar:     calendar(startByMonth, finish),
		}

		err = template.Execute(w, page)
		if err != nil {
			log.Println(err)
		}
	}
}
