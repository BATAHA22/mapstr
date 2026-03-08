package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Mistral struct {
	apiKey string
	model  string
}

func NewMistral() Provider {
	return &Mistral{
		apiKey: envKey("MISTRAL_API_KEY"),
		model:  "mistral-large-latest",
	}
}

func (m *Mistral) Name() string      { return "mistral" }
func (m *Mistral) Available() bool    { return m.apiKey != "" }
func (m *Mistral) SetModel(s string)  { m.model = s }

func (m *Mistral) Summarize(ctx context.Context, prompt string) (string, error) {
	body := map[string]any{
		"model": m.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 4096,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("mistral: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mistral.ai/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("mistral: request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("mistral: request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("mistral: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("mistral: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("mistral: unmarshal: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("mistral: empty response")
	}

	return result.Choices[0].Message.Content, nil
}
