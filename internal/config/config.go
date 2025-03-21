package config

import (
	"os"
)

const (
	OpenAI     = "openai"
	OpenRouter = "openrouter"
)

type Config struct {
	Provider         string
	APIKey           string
	Model            string
	MaxTokens        int
	Temperature      float64
	Timeout          int
	OutputDir        string
	ScenariosCount   int
	ValidateScenario int
	ValidateAll      bool
}

func NewConfig() *Config {
	return &Config{
		Provider:         OpenRouter,
		Model:            "",
		MaxTokens:        4000,
		Temperature:      0.7,
		OutputDir:        "scenarios",
		ScenariosCount:   5,
		ValidateScenario: 0,
		ValidateAll:      false,
	}
}

func (c *Config) GetDefaultModel() string {
	if c.Model != "" {
		return c.Model
	}

	switch c.Provider {
	case OpenAI:
		return "gpt-4"
	case OpenRouter:
		return "deepseek/deepseek-chat:free"
	default:
		return "deepseek/deepseek-chat:free"
	}
}

func (c *Config) GetProviderKey() string {
	if c.APIKey != "" {
		return c.APIKey
	}

	switch c.Provider {
	case OpenAI:
		return getEnv("OPENAI_API_KEY", "")
	case OpenRouter:
		return getEnv("OPENROUTER_API_KEY", "")
	default:
		return ""
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
