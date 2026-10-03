// Package scenario persists scenarios on disk.
//
// Each scenario lives in <dir>/scenario-NN/ and contains scenario.json (the
// source of truth) and scenario.md (a human-readable rendering).
package scenario

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/galalen/ckad-trainer/models"
)

const (
	dirPrefix       = "scenario-"
	jsonFile        = "scenario.json"
	markdownFile    = "scenario.md"
	LegacyScript    = "validate.sh"
	NamespaceMarker = "{{ns}}"
)

// Entry is a scenario directory found on disk.
type Entry struct {
	Index int
	Dir   string
}

// DirName returns the directory name for a scenario index.
func DirName(index int) string {
	return fmt.Sprintf("%s%02d", dirPrefix, index)
}

// List returns the scenario directories in dir, ordered by index.
func List(dir string) ([]Entry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var out []Entry
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), dirPrefix) {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(e.Name(), dirPrefix))
		if err != nil {
			continue
		}
		out = append(out, Entry{Index: n, Dir: filepath.Join(dir, e.Name())})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Index < out[j].Index })
	return out, nil
}

// NextIndex returns the first index after the highest existing scenario, so
// new scenarios never overwrite earlier ones.
func NextIndex(dir string) (int, error) {
	entries, err := List(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}
	if len(entries) == 0 {
		return 1, nil
	}
	return entries[len(entries)-1].Index + 1, nil
}

// NamespaceFor returns the dedicated namespace of a scenario index.
func NamespaceFor(index int) string {
	return fmt.Sprintf("ckad-%s", DirName(index))
}

// Save writes a scenario to <dir>/scenario-NN, replacing the {{ns}} marker
// with the scenario's dedicated namespace.
func Save(dir string, index int, s models.Scenario) (string, error) {
	s.Namespace = NamespaceFor(index)
	s = substituteNamespace(s)

	scenarioDir := filepath.Join(dir, DirName(index))
	if err := os.MkdirAll(scenarioDir, 0755); err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error encoding scenario: %w", err)
	}
	if err := os.WriteFile(filepath.Join(scenarioDir, jsonFile), data, 0644); err != nil {
		return "", fmt.Errorf("error writing scenario: %w", err)
	}
	if err := os.WriteFile(filepath.Join(scenarioDir, markdownFile), []byte(Markdown(s)), 0644); err != nil {
		return "", fmt.Errorf("error writing scenario description: %w", err)
	}
	return scenarioDir, nil
}

// Load reads scenario.json from a scenario directory.
func Load(scenarioDir string) (models.Scenario, error) {
	var s models.Scenario
	data, err := os.ReadFile(filepath.Join(scenarioDir, jsonFile))
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("error parsing %s: %w", jsonFile, err)
	}
	return s, nil
}

// HasJSON reports whether a scenario directory uses the per-task format.
func HasJSON(scenarioDir string) bool {
	_, err := os.Stat(filepath.Join(scenarioDir, jsonFile))
	return err == nil
}

// Markdown renders a scenario for reading. Hints and solutions are kept out
// so the file can be read while practising without spoilers.
func Markdown(s models.Scenario) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", s.Title)

	var meta []string
	if s.Domain != "" {
		meta = append(meta, "**Domain:** "+s.Domain)
	}
	if s.Difficulty != "" {
		meta = append(meta, "**Difficulty:** "+s.Difficulty)
	}
	if s.Namespace != "" {
		meta = append(meta, "**Namespace:** `"+s.Namespace+"`")
	}
	if len(meta) > 0 {
		b.WriteString(strings.Join(meta, " · ") + "\n\n")
	}

	fmt.Fprintf(&b, "## Description\n\n%s\n\n", s.Description)

	if len(s.Setup) > 0 {
		b.WriteString("## Setup\n\n```bash\n")
		for _, c := range s.Setup {
			b.WriteString(c + "\n")
		}
		b.WriteString("```\n\n")
	}

	b.WriteString("## Tasks\n\n")
	for i, t := range s.Tasks {
		fmt.Fprintf(&b, "%d. %s\n", i+1, t.Text)
	}

	if len(s.Cleanup) > 0 {
		b.WriteString("\n## Cleanup\n\n```bash\n")
		for _, c := range s.Cleanup {
			b.WriteString(c + "\n")
		}
		b.WriteString("```\n")
	}
	return b.String()
}

func substituteNamespace(s models.Scenario) models.Scenario {
	r := strings.NewReplacer(NamespaceMarker, s.Namespace)

	s.Title = r.Replace(s.Title)
	s.Description = r.Replace(s.Description)
	s.Setup = replaceAll(r, s.Setup)
	s.Cleanup = replaceAll(r, s.Cleanup)

	tasks := make([]models.Task, len(s.Tasks))
	for i, t := range s.Tasks {
		t.Text = r.Replace(t.Text)
		t.Hint = r.Replace(t.Hint)
		t.Solution = r.Replace(t.Solution)
		checks := make([]models.Check, len(t.Checks))
		for j, c := range t.Checks {
			c.Description = r.Replace(c.Description)
			c.Cmd = r.Replace(c.Cmd)
			c.Expect = r.Replace(c.Expect)
			checks[j] = c
		}
		t.Checks = checks
		tasks[i] = t
	}
	s.Tasks = tasks
	return s
}

func replaceAll(r *strings.Replacer, in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = r.Replace(s)
	}
	return out
}
