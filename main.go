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

	jsonCoverLetter, err := json.Marshal(coverLetter)
	if err != nil {
		log.Fatal(err)
	}

	var responseList []Response

	for _, job := range jobs {

		prompt := fmt.Sprintf("Cover Letter: %v\n\nJob history: %v\n\nJob Title: %v\n\nJob Description: %v",
			string(jsonCoverLetter),
			string(jsonHistory),
			job.Title,
			job.Description,
		)

		jsonResponse, err := askGPT(prompt)
		if err != nil {
			log.Fatal(err)
		}

		var response Response

		err = json.Unmarshal([]byte(jsonResponse), &response)
		if err != nil {
			log.Fatal(err)
		}

		if response.IsMatch {
			responseList = append(responseList, response)
		}
	}
	log.Println(responseList)
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

	c.OnHTML("a[aria-label='Next']", func(e *colly.HTMLElement) {
		nextPage := e.Attr("href")
		if nextPage != "" {
			e.Request.Visit("https://www.seek.com.au" + nextPage)
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
	rows, err := db.Query("SELECT name, role, start, finish, current, duties FROM history WHERE user_id = 1")

	if err != nil {
		return err
	}

	for rows.Next() {
		var history History
		err := rows.Scan(
			&history.Name,
			&history.Role,
			&history.Start,
			&history.Finish,
			&history.Current,
			&history.Duties,
		)
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
	err = db.QueryRow("SELECT content FROM letters WHERE user_id = 1").Scan(
		&coverLetter.Content,
	)

	if err != nil {
		return err
	}

	return nil
}

func askGPT(message string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	systemPrompt := `You are a job matching assistant. Your task is to evaluate whether a given job history matches the requirements of a job description. You will receive a JSON payload with a job history and a cover letter. Based on this information, you will determine if there is a match (isMatch: true or false). If isMatch is true, you will also provide a custom cover letter tailored to the job description.

The response should be in JSON format with the following structure:
{
  "isMatch": boolean,
  "coverLetter": string
}

If isMatch is false, the coverLetter should be an empty string.

Consider the following when making your decision:
- Relevance of job history to the job description
- Skills and experiences mentioned
- Any other relevant information

Make sure the cover letter is professional, concise, and highlights the candidate's strengths in relation to the job description.`

	requestBody := ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
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
