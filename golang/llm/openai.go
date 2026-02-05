package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// ChatCompletionRequest represents the request to OpenAI API
type ChatCompletionRequest struct {
	Model       string         `json:"model"`
	Messages    []ChatMessage  `json:"messages"`
	Temperature float32        `json:"temperature,omitempty"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	Stream      bool           `json:"stream"`
}

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents the response from OpenAI API
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int         `json:"index"`
		Message ChatMessage `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// APIError represents a structured error response from OpenAI
type APIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
	StatusCode int
}

// Client wraps OpenAI API interactions
type Client struct {
	apiKey  string
	model   string
	baseURL string
}

// NewClient creates a new OpenAI client
func NewClient() (*Client, error) {
	apiKey := os.Getenv("OPENAI_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_KEY environment variable not set")
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-5.2"
	}

	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.openai.com/v1",
	}, nil
}

// ChatCompletion sends a chat completion request to OpenAI
func (c *Client) ChatCompletion(messages []ChatMessage, temperature float32, maxTokens int) (string, error) {
	req := ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil {
			apiErr.StatusCode = resp.StatusCode
			return "", fmt.Errorf("OpenAI API error (%s): %s (HTTP %d)", apiErr.Error.Code, apiErr.Error.Message, resp.StatusCode)
		}
		return "", fmt.Errorf("OpenAI API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var completion ChatCompletionResponse
	if err := json.Unmarshal(respBody, &completion); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return completion.Choices[0].Message.Content, nil
}
