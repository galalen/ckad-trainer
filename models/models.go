package models

type LLMRequest struct {
	Model       string       `json:"model"`
	Messages    []LLMMessage `json:"messages"`
	MaxTokens   int          `json:"max_tokens"`
	Temperature float64      `json:"temperature"`
}

type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type ScenarioResponse struct {
	Scenarios []Scenario `json:"scenarios"`
}

type Scenario struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tasks       []string `json:"tasks"`
	Validation  string   `json:"validation"`
}
