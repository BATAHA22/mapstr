package llm

import "fmt"

// Usage holds token usage and cost info from an LLM API call.
type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	Provider     string
	Model        string
}

// Cost per 1M tokens [input, output] in USD.
var costPerMillion = map[string][2]float64{
	"claude":   {3.0, 15.0},   // Claude Sonnet 4.5
	"openai":   {2.50, 10.0},  // GPT-4o
	"gemini":   {1.25, 5.0},   // Gemini 2.5 Pro
	"deepseek": {0.14, 0.28},  // DeepSeek Chat
	"mistral":  {2.0, 6.0},    // Mistral Large
	"ollama":   {0, 0},        // Free (local)
}

// Cost returns the estimated API cost in USD.
func (u *Usage) Cost() float64 {
	if u == nil {
		return 0
	}
	rates, ok := costPerMillion[u.Provider]
	if !ok {
		return 0
	}
	return (float64(u.InputTokens)*rates[0] + float64(u.OutputTokens)*rates[1]) / 1_000_000
}

// CostString returns a human-readable cost string.
func (u *Usage) CostString() string {
	if u == nil {
		return "N/A"
	}
	cost := u.Cost()
	if cost == 0 {
		return "Free (local)"
	}
	return fmt.Sprintf("$%.4f (%s)", cost, u.Provider)
}
