package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/galalen/ckad-trainer/config"
	"github.com/galalen/ckad-trainer/generator"
	"github.com/galalen/ckad-trainer/validator"
)

func main() {
	cfg := config.NewConfig()

	rootCmd := &cobra.Command{
		Use:   "ckad-trainer",
		Short: "A CLI tool for generating CKAD practice scenarios",
		Long:  `CKAD Trainer generates Kubernetes certification scenarios and validation scripts using LLMs.`,
	}

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate CKAD scenarios",
		Run: func(cmd *cobra.Command, args []string) {
			if err := generator.GenerateScenarios(cfg); err != nil {
				color.Red("Error: %v", err)
				os.Exit(1)
			}
		},
	}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Run validation scripts for scenarios",
		Run: func(cmd *cobra.Command, args []string) {
			if err := validator.ValidateScenarios(cfg); err != nil {
				color.Red("Error: %v", err)
				os.Exit(1)
			}
		},
	}

	generateCmd.Flags().StringVarP(&cfg.Provider, "provider", "p", "openrouter", "LLM provider (openai or openrouter)")
	generateCmd.Flags().StringVarP(&cfg.APIKey, "api-key", "k", "", "API key for the chosen provider")
	generateCmd.Flags().StringVarP(&cfg.Model, "model", "m", "", "Model to use (default: gpt-4o-mini for openai, deepseek/deepseek-chat:free for openrouter)")
	generateCmd.Flags().StringVarP(&cfg.OutputDir, "output", "o", "scenarios", "Output directory for scenarios")
	generateCmd.Flags().IntVarP(&cfg.ScenariosCount, "count", "n", 5, "Number of scenarios to generate")
	generateCmd.Flags().IntVarP(&cfg.MaxTokens, "max-tokens", "", 4000, "Maximum tokens for response")
	generateCmd.Flags().Float64VarP(&cfg.Temperature, "temperature", "t", 0.7, "Temperature for generation")
	generateCmd.Flags().IntVarP(&cfg.Timeout, "timeout", "T", 2, "Request timeout (in minutes)")
	generateCmd.Flags().IntVarP(&cfg.Concurrency, "concurrency", "c", 3, "Number of scenarios generated in parallel")
	generateCmd.Flags().StringVar(&cfg.Topic, "topic", "", "Only generate scenarios for this skill area (e.g. ingress, jobs, secrets)")

	validateCmd.Flags().IntVarP(&cfg.ValidateScenario, "scenario", "s", 0, "Specific scenario number to validate (0 for none)")
	validateCmd.Flags().BoolVarP(&cfg.ValidateAll, "all", "a", false, "Validate all scenarios")
	validateCmd.Flags().StringVarP(&cfg.OutputDir, "dir", "d", "scenarios", "Directory containing scenarios")
	validateCmd.Flags().BoolVar(&cfg.AllowScripts, "allow-scripts", false, "Run legacy validate.sh scripts (arbitrary bash) for old scenarios")
	validateCmd.Flags().IntVar(&cfg.CheckTimeout, "check-timeout", 20, "Timeout for each check (in seconds)")

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
