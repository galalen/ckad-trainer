package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/galalen/ckad-trainer/checks"
	"github.com/galalen/ckad-trainer/config"
	"github.com/galalen/ckad-trainer/models"
)

const maxAttempts = 3

type Client struct {
	config *config.Config
	http   *http.Client
	// backoff is the delay before the first retry; it doubles per attempt.
	backoff time.Duration
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		config:  cfg,
		http:    &http.Client{Timeout: time.Duration(cfg.Timeout) * time.Minute},
		backoff: 2 * time.Second,
	}
}

// apiError is a non-200 response from the provider.
type apiError struct {
	status int
	body   string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("API error (%d): %s", e.status, e.body)
}

// retryable reports whether another attempt may succeed. Authentication and
// request errors will fail the same way every time.
func retryable(err error) bool {
	var ae *apiError
	if errors.As(err, &ae) {
		return ae.status == http.StatusTooManyRequests || ae.status >= 500
	}
	return !errors.Is(err, context.Canceled)
}

// GenerateScenario requests one scenario in the given domain, retrying on
// transient API failures and on malformed or invalid scenarios.
func (c *Client) GenerateScenario(ctx context.Context, domain string, avoid []string) (models.Scenario, error) {
	request := models.LLMRequest{
		Model: c.config.GetDefaultModel(),
		Messages: []models.LLMMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: BuildPrompt(domain, avoid)},
		},
		MaxTokens:      c.config.MaxTokens,
		Temperature:    c.config.Temperature,
		ResponseFormat: &models.ResponseFormat{Type: "json_object"},
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			delay := c.backoff << (attempt - 2)
			select {
			case <-ctx.Done():
				return models.Scenario{}, ctx.Err()
			case <-time.After(delay):
			}
		}

		content, err := c.callLLMAPI(ctx, request)
		if err == nil {
			var s models.Scenario
			s, err = parseScenario(content)
			if err == nil {
				return s, nil
			}
		}

		lastErr = err
		if !retryable(err) {
			break
		}
	}
	return models.Scenario{}, lastErr
}

func (c *Client) endpoint() string {
	switch c.config.Provider {
	case config.OpenAI:
		return "https://api.openai.com/v1/chat/completions"
	default:
		return "https://openrouter.ai/api/v1/chat/completions"
	}
}

func (c *Client) callLLMAPI(ctx context.Context, request models.LLMRequest) (string, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(), bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.GetProviderKey())

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", &apiError{status: resp.StatusCode, body: string(body)}
	}

	var response models.LLMResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no content in API response")
	}
	if response.Choices[0].FinishReason == "length" {
		return "", fmt.Errorf("response truncated at %d tokens; raise --max-tokens", c.config.MaxTokens)
	}

	return response.Choices[0].Message.Content, nil
}

func parseScenario(responseText string) (models.Scenario, error) {
	var s models.Scenario

	jsonContent := extractJSON(responseText)
	if jsonContent == "" {
		return s, fmt.Errorf("couldn't extract valid JSON from response")
	}

	if err := json.Unmarshal([]byte(jsonContent), &s); err != nil {
		return s, fmt.Errorf("failed to parse response: %w", err)
	}

	if strings.TrimSpace(s.Title) == "" {
		return s, fmt.Errorf("scenario has no title")
	}
	if err := checks.Validate(s); err != nil {
		return s, fmt.Errorf("invalid scenario %q: %w", s.Title, err)
	}
	return s, nil
}

// extractJSON returns the first complete top-level JSON object in text,
// tolerating markdown fences or prose around it.
func extractJSON(text string) string {
	start := strings.Index(text, "{")
	if start < 0 {
		return ""
	}

	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(text); i++ {
		ch := text[i]
		switch {
		case escaped:
			escaped = false
		case inString && ch == '\\':
			escaped = true
		case ch == '"':
			inString = !inString
		case inString:
		case ch == '{':
			depth++
		case ch == '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	return ""
}
