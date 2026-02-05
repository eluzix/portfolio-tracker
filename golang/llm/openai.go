package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"unicode"
)

// ResponseRequest represents the request to OpenAI Responses API
type ResponseRequest struct {
	Model               string           `json:"model"`
	Instructions        string           `json:"instructions,omitempty"`
	Input               any              `json:"input"`
	MaxOutputTokens     int              `json:"max_output_tokens,omitempty"`
	Reasoning           *ReasoningConfig `json:"reasoning,omitempty"`
	Store               bool             `json:"store"`
}

// ReasoningConfig configures reasoning behavior for supported models
type ReasoningConfig struct {
	Effort string `json:"effort"`
}

// InputMessage represents a single message in the input array
type InputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseOutput represents the response from OpenAI Responses API
type ResponseOutput struct {
	ID         string       `json:"id"`
	Object     string       `json:"object"`
	CreatedAt  int64        `json:"created_at"`
	Model      string       `json:"model"`
	Output     []OutputItem `json:"output"`
	OutputText string       `json:"output_text"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// OutputItem represents an item in the output array
type OutputItem struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Status  string `json:"status"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
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

// Generate sends a request to OpenAI Responses API
func (c *Client) Generate(instructions string, userInput string, maxTokens int) (string, error) {
	req := ResponseRequest{
		Model:           c.model,
		Instructions:    instructions,
		Input:           userInput,
		MaxOutputTokens: maxTokens,
		Reasoning:       &ReasoningConfig{Effort: "low"},
		Store:           false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/responses", bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil {
			apiErr.StatusCode = resp.StatusCode
			return "", fmt.Errorf("OpenAI API error (%s): %s (HTTP %d)", apiErr.Error.Code, apiErr.Error.Message, resp.StatusCode)
		}
		return "", fmt.Errorf("OpenAI API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var response ResponseOutput
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if response.OutputText != "" {
		return response.OutputText, nil
	}

	for _, item := range response.Output {
		if item.Type == "message" && len(item.Content) > 0 {
			for _, content := range item.Content {
				if content.Type == "output_text" || content.Type == "text" {
					return content.Text, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no text content in response")
}

func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteByte('\n')
		}
		result.WriteString(wrapLine(line, width))
	}

	return result.String()
}

func wrapLine(line string, width int) string {
	if len(line) <= width {
		return line
	}

	var result strings.Builder
	words := strings.Fields(line)
	if len(words) == 0 {
		return line
	}

	currentLen := 0
	for i, word := range words {
		wordLen := len(word)

		if i == 0 {
			result.WriteString(word)
			currentLen = wordLen
			continue
		}

		if currentLen+1+wordLen > width {
			result.WriteByte('\n')
			leadingSpaces := countLeadingSpaces(line)
			if leadingSpaces > 0 && leadingSpaces < width/2 {
				result.WriteString(strings.Repeat(" ", leadingSpaces))
				currentLen = leadingSpaces
			} else {
				currentLen = 0
			}
			result.WriteString(word)
			currentLen += wordLen
		} else {
			result.WriteByte(' ')
			result.WriteString(word)
			currentLen += 1 + wordLen
		}
	}

	return result.String()
}

func countLeadingSpaces(s string) int {
	count := 0
	for _, r := range s {
		if unicode.IsSpace(r) {
			count++
		} else {
			break
		}
	}
	return count
}
