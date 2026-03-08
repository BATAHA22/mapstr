package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Ollama struct {
	baseURL string
	model   string
}

func NewOllama() Provider {
	baseURL := envKey("OLLAMA_HOST")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &Ollama{
		baseURL: baseURL,
		model:   "llama3",
	}
}

func (o *Ollama) Name() string      { return "ollama" }
func (o *Ollama) SetModel(m string)  { o.model = m }

func (o *Ollama) Available() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(o.baseURL + "/api/tags")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (o *Ollama) Summarize(ctx context.Context, prompt string) (string, *Usage, error) {
	body := map[string]any{
		"model":  o.model,
		"prompt": prompt,
		"stream": false,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", nil, fmt.Errorf("ollama: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewReader(payload))
	if err != nil {
		return "", nil, fmt.Errorf("ollama: request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Ollama can be slow with large prompts, use a longer timeout.
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("ollama: request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("ollama: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("ollama: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Response        string `json:"response"`
		PromptEvalCount int    `json:"prompt_eval_count"`
		EvalCount       int    `json:"eval_count"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", nil, fmt.Errorf("ollama: unmarshal: %w", err)
	}

	if result.Response == "" {
		return "", nil, fmt.Errorf("ollama: empty response")
	}

	usage := &Usage{
		InputTokens:  result.PromptEvalCount,
		OutputTokens: result.EvalCount,
		TotalTokens:  result.PromptEvalCount + result.EvalCount,
		Provider:     "ollama",
		Model:        o.model,
	}

	return result.Response, usage, nil
}
