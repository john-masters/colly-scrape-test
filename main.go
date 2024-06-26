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
	searchTerms := []string{"full stack developer", "backend developer", "frontend developer"}

	var jobDetails []Job

	getJobs(&jobDetails, searchTerms)

	log.Println(jobDetails)
}

func getJobs(jobDetails *[]Job, searchTerms []string) {
	var wg sync.WaitGroup

	for _, term := range searchTerms {
		wg.Add(1)

		go func(term string) {
			defer wg.Done()

			var pageUrls []string
			getPageUrls(&pageUrls, term)

			var jobUrls []string
			getJobUrls(&pageUrls, &jobUrls)

			getJobDetails(&jobUrls, jobDetails)

		}(term)

		wg.Wait()
	}
}

func getPageUrls(pages *[]string, searchTerm string) {
	c := colly.NewCollector()

	c.OnHTML("a[aria-label='Next']", func(e *colly.HTMLElement) {
		*pages = append(*pages, e.Request.URL.String())

		nextPage := e.Attr("href")
		if nextPage != "" {
			e.Request.Visit("https://www.seek.com.au" + nextPage)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	formattedSearchTerm := strings.ReplaceAll(searchTerm, " ", "-")
	url := "https://www.seek.com.au/" + formattedSearchTerm + "-jobs/full-time?daterange=1"
	c.Visit(url)
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

			c.OnRequest(func(r *colly.Request) {
				log.Println("Visiting", r.URL.String())
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

			c.OnRequest(func(r *colly.Request) {
				log.Println("Visiting", r.URL.String())
			})

			c.Visit(url)
		}(url)

		wg.Wait()
	}
}
