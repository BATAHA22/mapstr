package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DeepSeek struct {
	apiKey string
	model  string
}

func NewDeepSeek() Provider {
	return &DeepSeek{
		apiKey: envKey("DEEPSEEK_API_KEY"),
		model:  "deepseek-chat",
	}
}

func (d *DeepSeek) Name() string      { return "deepseek" }
func (d *DeepSeek) Available() bool    { return d.apiKey != "" }
func (d *DeepSeek) SetModel(m string)  { d.model = m }

func (d *DeepSeek) Summarize(ctx context.Context, prompt string) (string, *Usage, error) {
	body := map[string]any{
		"model": d.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 4096,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", nil, fmt.Errorf("deepseek: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.deepseek.com/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", nil, fmt.Errorf("deepseek: request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("deepseek: request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("deepseek: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("deepseek: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", nil, fmt.Errorf("deepseek: unmarshal: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", nil, fmt.Errorf("deepseek: empty response")
	}

	usage := &Usage{
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
		TotalTokens:  result.Usage.TotalTokens,
		Provider:     "deepseek",
		Model:        d.model,
	}

	return result.Choices[0].Message.Content, usage, nil
}
