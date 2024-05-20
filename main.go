package main

import (
	"log"

	"github.com/gocolly/colly"
)

func scrape(jobs *[]Job) {
	c := colly.NewCollector()

	c.OnHTML("[data-card-type='JobCard']", func(e *colly.HTMLElement) {
		title := e.ChildText("a[data-automation='jobTitle']")
		company := e.ChildText("a[data-automation='jobCompany']")
		link := e.ChildAttr("a[data-automation='jobTitle']", "href")

		fullLink := "https://www.seek.com.au" + link

		job := Job{
			Title:   title,
			Company: company,
			Link:    fullLink,
		}

		*jobs = append(*jobs, job)

		e.Request.Visit(fullLink)
	})

	c.OnHTML("div[data-automation='jobAdDetails']", func(e *colly.HTMLElement) {
		description := e.Text

		for i := range *jobs {
			if (*jobs)[i].Link == e.Request.URL.String() {
				(*jobs)[i].Description = description
				break
			}
		}
	})

	c.Visit("https://www.seek.com.au/junior-developer-jobs/full-time?daterange=1")
}

func main() {
	var jobs []Job

	scrape(&jobs)

	log.Println("Jobs found: ", len(jobs))

	for _, job := range jobs {
		log.Println("Title: ", job.Title)
		log.Println("Company: ", job.Company)
		log.Println("Link: ", job.Link)
		log.Println("Description: ", job.Description)
	}
}
