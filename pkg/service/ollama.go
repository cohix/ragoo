package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type ollamaService struct {
	config map[string]string
}

type ollamaRequest struct {
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
	Messages []message `json:"messages"`
}

type ollamaResponse struct {
	Message message `json:"message"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (o *ollamaService) Completion(prompt string) (*Result, error) {
	model, exists := o.config["model"]
	if !exists {
		return nil, errors.New("ollama service missing config key: model")
	}

	url := "http://localhost:11434/api/chat"

	reqBody := &ollamaRequest{
		Model:  model,
		Stream: false,
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to json.Marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to NewRequest: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to Do: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	respBody := &ollamaResponse{}
	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return nil, fmt.Errorf("failed to NewDecoder.Decode: %w", err)
	}

	r := &Result{
		Completion: respBody.Message.Content,
	}

	return r, nil
}
