package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAI struct {
	apiKey  string
	model   string
	baseURL string
}

func NewOpenAI() Provider {
	baseURL := envKey("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAI{
		apiKey:  envKey("OPENAI_API_KEY"),
		model:   "gpt-4o",
		baseURL: baseURL,
	}
}

func (o *OpenAI) Name() string      { return "openai" }
func (o *OpenAI) Available() bool    { return o.apiKey != "" }
func (o *OpenAI) SetModel(m string)  { o.model = m }

func (o *OpenAI) Summarize(ctx context.Context, prompt string) (string, error) {
	body := map[string]any{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 4096,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("openai: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("openai: request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai: request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("openai: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("openai: unmarshal: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openai: empty response")
	}

	return result.Choices[0].Message.Content, nil
}
