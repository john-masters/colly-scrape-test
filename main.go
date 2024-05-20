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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var jobs []Job
	var historyList []History
	var coverLetter Letter

	scrape(&jobs)
	err = getHistory(&historyList)
	if err != nil {
		log.Fatal(err)
	}

	err = getCoverLetter(&coverLetter)
	if err != nil {
		log.Fatal(err)
	}

	jsonHistory, err := json.Marshal(historyList)
	if err != nil {
		log.Fatal(err)
	}

	for i, job := range jobs {
		if i > 0 { // loop once for testing
			break
		}

		prompt := fmt.Sprintf(`
			Do you think I'd be a good fit for this role?
			Please compare my cover letter and job history
			to the job description and let me know
			if you think I should apply.

			Cover Letter: %v
			Job history: %v


			Job Title: %v
			Job Description: %v`,
			coverLetter.Content,
			string(jsonHistory),
			job.Title,
			job.Description,
		)

		log.Println("Prompt: ", prompt)

		response, err := askGPT(prompt)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Response: ", response)

		// 	log.Println("Title: ", job.Title)
		// 	log.Println("Company: ", job.Company)
		// 	log.Println("Link: ", job.Link)
		// 	log.Println("Description: ", job.Description)

		// 	response, err := askGPT(job.Description)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}

		// 	log.Println("Response: ", response)

	}
}

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

func dbConnect() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./test.db")

	if err != nil {
		return db, err
	}

	// don't defer here as this will close when function returns
	// defer db.Close()

	return db, err
}

func getHistory(historyList *[]History) error {
	db, err := dbConnect()
	if err != nil {
		return err
	}
	defer db.Close()

	// hardcoded id for testing
	rows, err := db.Query("SELECT * FROM history WHERE user_id = 1")

	if err != nil {
		return err
	}

	for rows.Next() {
		var history History
		err := rows.Scan(&history.ID, &history.UserID, &history.Name, &history.Role, &history.Start, &history.Finish, &history.Current, &history.Duties)
		if err != nil {
			return err
		}

		*historyList = append(*historyList, history)

	}

	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func getCoverLetter(coverLetter *Letter) error {
	db, err := dbConnect()
	if err != nil {
		return err
	}
	defer db.Close()

	// hardcoded id for testing
	err = db.QueryRow("SELECT * FROM letters WHERE user_id = 1").Scan(
		&coverLetter.ID,
		&coverLetter.UserID,
		&coverLetter.Content,
		&coverLetter.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
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
