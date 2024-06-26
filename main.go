package main

import (
	"log"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

type Job struct {
	Position    string
	Company     string
	Location    string
	Description string
	Link        string
}

func main() {
	var pageUrls []string
	getPageUrls(&pageUrls)

	var jobUrls []string
	getJobUrls(&pageUrls, &jobUrls)

	var jobDetails []Job
	getJobDetails(&jobUrls, &jobDetails)

	log.Println(jobDetails)
}

func getPageUrls(pages *[]string) {
	c := colly.NewCollector()

	c.OnHTML("a[aria-label='Next']", func(e *colly.HTMLElement) {
		*pages = append(*pages, e.Request.URL.String())

		nextPage := e.Attr("href")
		if nextPage != "" {
			e.Request.Visit("https://www.seek.com.au" + nextPage)
		}
	})

	c.Visit("https://www.seek.com.au/full-stack-developer-jobs/full-time?daterange=1")
}

func getJobUrls(pageUrls *[]string, jobUrls *[]string) {

	var wg sync.WaitGroup

	for _, url := range *pageUrls {
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			c := colly.NewCollector()

			c.OnHTML("[data-automation='normalJob']", func(e *colly.HTMLElement) {
				route := e.ChildAttr("a[data-automation='jobTitle']", "href")
				link := "https://www.seek.com.au" + route

				*jobUrls = append(*jobUrls, link)
			})

			c.Visit(url)
		}(url)

		wg.Wait()
	}
}

func getJobDetails(jobUrls *[]string, jobDetails *[]Job) {

	var wg sync.WaitGroup

	for _, url := range *jobUrls {
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			c := colly.NewCollector()

			c.OnHTML("div[data-automation='jobDetailsPage']", func(e *colly.HTMLElement) {
				position := e.ChildText("[data-automation='job-detail-title']")
				company := e.ChildText("[data-automation='advertiser-name']")
				location := e.ChildText("[data-automation='job-detail-location']")
				description := e.ChildText("[data-automation='jobAdDetails']")
				formattedDescription := strings.ReplaceAll(description, "\n", "\\n")

				job := Job{
					Position:    position,
					Company:     company,
					Location:    location,
					Description: formattedDescription,
					Link:        e.Request.URL.String(),
				}

				*jobDetails = append(*jobDetails, job)
			})

			c.Visit(url)
		}(url)

		wg.Wait()
	}
}
