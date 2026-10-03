package generator

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/galalen/ckad-trainer/config"
	"github.com/galalen/ckad-trainer/llm"
	"github.com/galalen/ckad-trainer/scenario"
)

func GenerateScenarios(cfg *config.Config) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	nextIndex, err := scenario.NextIndex(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("error reading output directory: %w", err)
	}

	domains, err := pickDomains(cfg.Topic, cfg.ScenariosCount)
	if err != nil {
		return err
	}
	avoid := existingTitles(cfg.OutputDir)

	client := llm.NewClient(cfg)
	ctx := context.Background()

	color.Green("Generating %d CKAD scenarios using %s / %s...", cfg.ScenariosCount, cfg.Provider, cfg.GetDefaultModel())

	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		created  int
		failures []string
		sem      = make(chan struct{}, max(cfg.Concurrency, 1))
	)

	for i, domain := range domains {
		wg.Add(1)
		go func(n int, domain string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			s, err := client.GenerateScenario(ctx, domain, avoid)

			// Indices are assigned in completion order so failed requests
			// leave no gaps between scenario directories.
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failures = append(failures, domain)
				color.Red("✗ [%d/%d] %s: %v", n, len(domains), domain, err)
				return
			}
			if _, err := scenario.Save(cfg.OutputDir, nextIndex, s); err != nil {
				failures = append(failures, domain)
				color.Red("✗ [%d/%d] %s: %v", n, len(domains), domain, err)
				return
			}
			color.Green("✓ [%d/%d] Created %s: %s", n, len(domains), scenario.DirName(nextIndex), s.Title)
			nextIndex++
			created++
		}(i+1, domain)
	}
	wg.Wait()

	if created == 0 {
		return fmt.Errorf("no scenarios were generated")
	}
	color.Green("Generated %d scenarios in %s", created, cfg.OutputDir)
	if len(failures) > 0 {
		color.Yellow("%d scenarios failed (%s); run generate again to fill the gap", len(failures), strings.Join(failures, ", "))
	}
	return nil
}

func validateConfig(cfg *config.Config) error {
	if cfg.Provider != config.OpenAI && cfg.Provider != config.OpenRouter {
		return fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}

	if cfg.GetProviderKey() == "" {
		return fmt.Errorf("no API key provided for %s provider", cfg.Provider)
	}

	if cfg.ScenariosCount < 1 {
		return fmt.Errorf("--count must be at least 1")
	}

	return nil
}

// pickDomains returns one domain per scenario. With a topic, every scenario
// uses the matching domain; otherwise domains rotate from a random start so
// repeated runs cover different areas.
func pickDomains(topic string, count int) ([]string, error) {
	if topic != "" {
		for _, d := range llm.Domains {
			if strings.Contains(strings.ToLower(d), strings.ToLower(topic)) {
				return repeat(d, count), nil
			}
		}
		return nil, fmt.Errorf("unknown topic %q; choose one of: %s", topic, strings.Join(llm.Domains, ", "))
	}

	start := rand.Intn(len(llm.Domains))
	out := make([]string, count)
	for i := range out {
		out[i] = llm.Domains[(start+i)%len(llm.Domains)]
	}
	return out, nil
}

func repeat(s string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = s
	}
	return out
}

func existingTitles(dir string) []string {
	entries, err := scenario.List(dir)
	if err != nil {
		return nil
	}
	var titles []string
	for _, e := range entries {
		if s, err := scenario.Load(e.Dir); err == nil && s.Title != "" {
			titles = append(titles, s.Title)
		}
	}
	return titles
}
