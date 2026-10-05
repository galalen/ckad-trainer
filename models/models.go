package models

type LLMRequest struct {
	Model          string          `json:"model"`
	Messages       []LLMMessage    `json:"messages"`
	MaxTokens      int             `json:"max_tokens"`
	Temperature    float64         `json:"temperature"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

type ResponseFormat struct {
	Type string `json:"type"`
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
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// Scenario is a single CKAD practice exercise. It is stored as scenario.json
// inside its scenario directory and is the source of truth for validation.
type Scenario struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Domain      string   `json:"domain"`
	Difficulty  string   `json:"difficulty"`
	Namespace   string   `json:"namespace"`
	Setup       []string `json:"setup"`
	Cleanup     []string `json:"cleanup"`
	Tasks       []Task   `json:"tasks"`
}

// Task is one step of a scenario, verified independently by its checks.
type Task struct {
	Text     string  `json:"text"`
	Hint     string  `json:"hint"`
	Solution string  `json:"solution"`
	Checks   []Check `json:"checks"`
}

// Check is a read-only kubectl command whose output is compared to Expect.
type Check struct {
	Description string `json:"description"`
	Cmd         string `json:"cmd"`
	Expect      string `json:"expect"`
	// Match is one of: equals (default when Expect is set), contains, regex,
	// exit0 (default when Expect is empty).
	Match string `json:"match"`
}
