package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/galalen/ckad-trainer/internal/config"
	"github.com/galalen/ckad-trainer/internal/llm"
	"github.com/galalen/ckad-trainer/internal/models"
)

func GenerateScenarios(cfg *config.Config) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	client := llm.NewClient(cfg)

	color.Green("Generating %d CKAD scenarios using %s / %s...", cfg.ScenariosCount, cfg.Provider, cfg.Model)
	scenarios, err := client.GenerateScenarios()
	if err != nil {
		return fmt.Errorf("error generating scenarios: %w", err)
	}

	for i, scenario := range scenarios {
		if err := saveScenario(cfg.OutputDir, i+1, scenario); err != nil {
			return fmt.Errorf("error saving scenario %d: %w", i+1, err)
		}
	}

	color.Green("Successfully generated %d scenarios in %s", len(scenarios), cfg.OutputDir)
	return nil
}

func validateConfig(cfg *config.Config) error {
	if cfg.Provider != config.OpenAI && cfg.Provider != config.OpenRouter {
		return fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}

	if cfg.GetProviderKey() == "" {
		return fmt.Errorf("no API key provided for %s provider", cfg.Provider)
	}

	return nil
}

func saveScenario(outputDir string, index int, scenario models.Scenario) error {
	scenarioDir := filepath.Join(outputDir, fmt.Sprintf("scenario-%02d", index))
	if err := os.MkdirAll(scenarioDir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	scenarioMD := fmt.Sprintf("# %s\n\n## Description\n\n%s\n\n## Tasks\n\n",
		scenario.Title, scenario.Description)

	for i, task := range scenario.Tasks {
		scenarioMD += fmt.Sprintf("%d. %s\n", i+1, task)
	}

	if err := os.WriteFile(filepath.Join(scenarioDir, "scenario.md"), []byte(scenarioMD), 0644); err != nil {
		return fmt.Errorf("error writing scenario description: %w", err)
	}

	if err := os.WriteFile(filepath.Join(scenarioDir, "validate.sh"), []byte(scenario.Validation), 0755); err != nil {
		return fmt.Errorf("error writing validation script: %w", err)
	}

	color.Green("✓ Created scenario %d: %s", index, scenario.Title)
	return nil
}
