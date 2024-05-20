package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var jobs []Job

	scrape(&jobs)

	for i, job := range jobs {
		if i > 0 { // loop once for testing
			break
		}

		log.Println("Title: ", job.Title)
		log.Println("Company: ", job.Company)
		log.Println("Link: ", job.Link)
		log.Println("Description: ", job.Description)

		response, err := askGPT(job.Description)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Response: ", response)

	}
}

func dbConnect() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./test.db")

	if err != nil {
		fmt.Println("Error opening database")
		return db, err
	}

	defer db.Close()

	return db, err
}

func askGPT(message string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"
	requestBody := ChatCompletionRequest{
		Model: "gpt-4o-2024-05-13",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: message},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	var response ChatCompletionResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	return response.Choices[0].Message.Content, nil
}
