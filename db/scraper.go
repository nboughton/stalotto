package db

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/nboughton/stalotto/lotto"
)

var (
	baseURL    = "https://www.lottery.co.uk"
	archiveURL = "%s/lotto/results/archive-%d"
)

// Scrape archive for data
func Scrape() <-chan lotto.Result {
	c := make(chan lotto.Result)

	go func() {
		defer close(c)

		for year := time.Now().Year(); year >= 1994; year-- {
			// Get archive page
			archivePage, err := goquery.NewDocument(fmt.Sprintf(archiveURL, baseURL, year))
			if err != nil {
				log.Println(err)
				break
			}

			// Find all results pages linked from archive page
			archivePage.Find("#siteContainer .main .lotto tbody tr td a").Each(func(i int, s *goquery.Selection) {
				resultURL, ok := s.Attr("href")
				if !ok {
					log.Println("No result URL for", s.Text())
					return
				}

				res, err := parseResultPage(resultURL)
				if err != nil {
					log.Println(err)
					return
				}

				c <- res
			})
		}
	}()

	return c
}

func parseResultPage(url string) (lotto.Result, error) {
	// Create new lotto.Result
	res := lotto.NewResult()

	// Load results page
	resultPage, err := goquery.NewDocument(fmt.Sprintf("%s%s", baseURL, url))
	if err != nil {
		log.Println(err)
		return res, err
	}

	// Set lotto.Result date
	if res.Date, err = parseDateFromURL(url); err != nil {
		log.Println(err)
		return res, err
	}

	// Set lotto.Result ball results
	resultPage.Find(".result").Each(func(i int, s *goquery.Selection) {
		result, err := strconv.Atoi(s.Text())
		if err != nil {
			log.Println(err)
		}

		if i < len(res.Balls) {
			res.Balls[i] = result
		} else {
			res.Bonus = result
		}
	})

	// Set lotto.Result machine and set
	resultPage.Find("#siteContainer .main .lotto tbody tr td").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Set Used:") {
			n, err := strconv.Atoi(parseUsed(s.Text()))
			if err != nil {
				log.Println(err)
			}

			res.Set = n
		}

		if strings.Contains(s.Text(), "Machine Used:") {
			res.Machine = parseUsed(s.Text())
		}
	})

	return res, nil
}

func parseUsed(str string) string {
	return strings.TrimSpace(strings.Split(str, ":")[1])
}

func parseDateFromURL(url string) (time.Time, error) {
	s := strings.Split(url, "s-")
	if len(s) == 2 {
		return time.Parse("02-01-2006", s[1])
	}

	return time.Now(), fmt.Errorf("bad url: %s", url)
}
