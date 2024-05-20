package main

import "time"

type Job struct {
	Title       string
	Company     string
	Link        string
	Description string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type History struct {
	ID      int
	UserID  int
	Name    string
	Role    string
	Start   string
	Finish  string
	Current bool
	Duties  string
}

type Letter struct {
	ID        int
	UserID    int
	Content   string
	CreatedAt time.Time
}
