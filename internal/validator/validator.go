package validator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/galalen/ckad-trainer/internal/config"
)

func ValidateScenarios(cfg *config.Config) error {
	if !cfg.ValidateAll && cfg.ValidateScenario == 0 {
		return fmt.Errorf("must specify --scenario or --all flag")
	}

	if !fileExists(cfg.OutputDir) {
		return fmt.Errorf("scenarios directory %s does not exist", cfg.OutputDir)
	}

	if cfg.ValidateAll {
		entries, err := os.ReadDir(cfg.OutputDir)
		if err != nil {
			return fmt.Errorf("error reading scenarios directory: %w", err)
		}

		failures := 0
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "scenario-") {
				numStr := strings.TrimPrefix(entry.Name(), "scenario-")
				if err := validateScenario(cfg.OutputDir, numStr); err != nil {
					color.Red("Error validating scenario %s: %v", numStr, err)
					failures++
				}
			}
		}

		if failures > 0 {
			return fmt.Errorf("%d scenarios failed validation", failures)
		}

		color.Green("All scenarios validated successfully")
		return nil
	} else {
		scenarioNum := fmt.Sprintf("%02d", cfg.ValidateScenario)
		return validateScenario(cfg.OutputDir, scenarioNum)
	}
}

func validateScenario(outputDir, scenarioNum string) error {
	scenarioDir := filepath.Join(outputDir, fmt.Sprintf("scenario-%s", scenarioNum))
	validateScript := filepath.Join(scenarioDir, "validate.sh")

	if !fileExists(validateScript) {
		return fmt.Errorf("validation script for scenario %s does not exist", scenarioNum)
	}

	if err := os.Chmod(validateScript, 0755); err != nil {
		return fmt.Errorf("error making validation script executable: %w", err)
	}

	color.Yellow("Running validation for scenario %s...", scenarioNum)

	cmd := exec.Command("/bin/bash", validateScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	color.Green("Scenario %s validated successfully", scenarioNum)
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
