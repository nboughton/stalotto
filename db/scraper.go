package db

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	baseURL    = "https://www.lottery.co.uk"
	archiveURL = "%s/lotto/results/archive-%d"
)

// Scrape archive for data
func Scrape() <-chan Record {
	c := make(chan Record)

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

				rec, err := parseResultPage(resultURL)
				if err != nil {
					log.Println(err)
					return
				}

				c <- rec
			})
		}
	}()

	return c
}

func parseResultPage(url string) (Record, error) {
	// Create new Record
	rec := NewRecord()

	// Load results page
	resultPage, err := goquery.NewDocument(fmt.Sprintf("%s%s", baseURL, url))
	if err != nil {
		log.Println(err)
		return rec, err
	}

	// Set Record date
	if rec.Date, err = parseDateFromURL(url); err != nil {
		log.Println(err)
		return rec, err
	}

	// Set Record ball results
	resultPage.Find(".result").Each(func(i int, s *goquery.Selection) {
		result, err := strconv.Atoi(s.Text())
		if err != nil {
			log.Println(err)
		}

		if i < len(rec.Ball) {
			rec.Ball[i] = result
		}
	})

	// Set Record machine and set
	resultPage.Find("#siteContainer .main .lotto tbody tr td").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Set Used:") {
			n, err := strconv.Atoi(parseUsed(s.Text()))
			if err != nil {
				log.Println(err)
			}

			rec.Set = n
		}

		if strings.Contains(s.Text(), "Machine Used:") {
			rec.Machine = parseUsed(s.Text())
		}
	})

	return rec, nil
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
