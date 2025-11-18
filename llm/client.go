package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/galalen/ckad-trainer/config"
	"github.com/galalen/ckad-trainer/models"
)

type Client struct {
	config *config.Config
	http   *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		http:   &http.Client{Timeout: time.Duration(cfg.Timeout) * time.Minute},
	}
}

func (c *Client) GenerateScenarios() ([]models.Scenario, error) {
	prompt := BuildPrompt(c.config.ScenariosCount)

	endpoint, headers := c.getProviderConfig()

	request := models.LLMRequest{
		Model: c.config.GetDefaultModel(),
		Messages: []models.LLMMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	}

	jsonContent, err := c.callLLMAPI(endpoint, headers, request)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	return c.parseResponse(jsonContent)
}

func (c *Client) getProviderConfig() (string, map[string]string) {
	var endpoint string
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + c.config.GetProviderKey()

	switch c.config.Provider {
	case config.OpenAI:
		endpoint = "https://api.openai.com/v1/chat/completions"
	case config.OpenRouter:
		endpoint = "https://openrouter.ai/api/v1/chat/completions"
	}

	return endpoint, headers
}

func (c *Client) callLLMAPI(endpoint string, headers map[string]string, request models.LLMRequest) (string, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

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
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var response models.LLMResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no content in API response")
	}

	return response.Choices[0].Message.Content, nil
}

func (c *Client) parseResponse(responseText string) ([]models.Scenario, error) {
	jsonContent := c.extractJSON(responseText)
	if jsonContent == "" {
		return nil, fmt.Errorf("couldn't extract valid JSON from response")
	}

	var scenarioResponse models.ScenarioResponse
	if err := json.Unmarshal([]byte(jsonContent), &scenarioResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return scenarioResponse.Scenarios, nil
}

func (c *Client) extractJSON(text string) string {
	startIdx := strings.Index(text, "{")
	endIdx := strings.LastIndex(text, "}")

	if startIdx >= 0 && endIdx > startIdx {
		return text[startIdx : endIdx+1]
	}

	return ""
}
