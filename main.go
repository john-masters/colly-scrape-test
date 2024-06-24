package main

import (
	"log"

	"github.com/gocolly/colly"
)

func main() {
	var pageUrls []string

	getPageUrls(&pageUrls)
	log.Println(pageUrls)

	// var jobs []Job

	// scrape(&jobs)

	// // loop over jobs
	// file, err := os.Create("result.csv")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// w := csv.NewWriter(file)
	// defer w.Flush()

	// var jobData [][]string
	// for _, job := range jobs {
	// 	jobData = append(jobData, []string{job.Title, job.Company, job.Link, job.Description})
	// }

	// w.WriteAll(jobData)

	// if err != nil {
	// 	log.Fatal(err)
	// }
}

// func scrape(jobs *[]Job) {
// 	c := colly.NewCollector()

// 	c.OnHTML("[data-card-type='JobCard']", func(e *colly.HTMLElement) {
// 		title := e.ChildText("a[data-automation='jobTitle']")
// 		company := e.ChildText("a[data-automation='jobCompany']")
// 		link := e.ChildAttr("a[data-automation='jobTitle']", "href")

// 		fullLink := "https://www.seek.com.au" + link

// 		job := Job{
// 			Title:   title,
// 			Company: company,
// 			Link:    fullLink,
// 		}

// 		*jobs = append(*jobs, job)

// 		e.Request.Visit(fullLink)
// 	})

// 	c.OnHTML("div[data-automation='jobAdDetails']", func(e *colly.HTMLElement) {
// 		description := e.Text

// 		for i := range *jobs {
// 			if (*jobs)[i].Link == e.Request.URL.String() {
// 				(*jobs)[i].Description = description
// 				break
// 			}
// 		}
// 	})

// 	c.OnHTML("a[aria-label='Next']", func(e *colly.HTMLElement) {
// 		nextPage := e.Attr("href")
// 		if nextPage != "" {
// 			e.Request.Visit("https://www.seek.com.au" + nextPage)
// 		}
// 	})

// 	c.OnRequest(func(r *colly.Request) {
// 		log.Println("Visiting", r.URL)
// 	})

// 	c.Visit("https://www.seek.com.au/full-stack-developer-jobs/full-time?daterange=1")
// }

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

