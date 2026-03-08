package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Claude struct {
	apiKey string
	model  string
}

func NewClaude() Provider {
	return &Claude{
		apiKey: envKey("ANTHROPIC_API_KEY"),
		model:  "claude-sonnet-4-5-20250514",
	}
}

func (c *Claude) Name() string      { return "claude" }
func (c *Claude) Available() bool    { return c.apiKey != "" }
func (c *Claude) SetModel(m string)  { c.model = m }

func (c *Claude) Summarize(ctx context.Context, prompt string) (string, error) {
	body := map[string]any{
		"model":      c.model,
		"max_tokens": 4096,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("claude: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("claude: request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude: request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("claude: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("claude: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("claude: unmarshal: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("claude: empty response")
	}

	return result.Content[0].Text, nil
}
