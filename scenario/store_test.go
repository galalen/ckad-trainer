package scenario

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/galalen/ckad-trainer/models"
)

func TestSaveLoadAndNextIndex(t *testing.T) {
	dir := t.TempDir()

	if n, err := NextIndex(filepath.Join(dir, "missing")); err != nil || n != 1 {
		t.Fatalf("NextIndex(missing) = %d, %v; want 1, nil", n, err)
	}

	s := models.Scenario{
		Title:   "Web",
		Setup:   []string{"kubectl create namespace {{ns}}"},
		Cleanup: []string{"kubectl delete namespace {{ns}}"},
		Tasks: []models.Task{{
			Text:     "Create deployment web in {{ns}}",
			Solution: "kubectl create deploy web --image=nginx:1.25 -n {{ns}}",
			Checks:   []models.Check{{Cmd: "kubectl get deploy web -n {{ns}}"}},
		}},
	}

	scenarioDir, err := Save(dir, 3, s)
	if err != nil {
		t.Fatal(err)
	}

	got, err := Load(scenarioDir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Namespace != "ckad-scenario-03" {
		t.Errorf("Namespace = %q", got.Namespace)
	}
	if got.Tasks[0].Checks[0].Cmd != "kubectl get deploy web -n ckad-scenario-03" {
		t.Errorf("check not substituted: %q", got.Tasks[0].Checks[0].Cmd)
	}
	if got.Setup[0] != "kubectl create namespace ckad-scenario-03" {
		t.Errorf("setup not substituted: %q", got.Setup[0])
	}
	if s.Tasks[0].Checks[0].Cmd != "kubectl get deploy web -n {{ns}}" {
		t.Errorf("Save mutated its input")
	}

	md, err := os.ReadFile(filepath.Join(scenarioDir, "scenario.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(md), "--image=nginx") {
		t.Errorf("scenario.md leaks the solution")
	}

	// A non-scenario directory and a gap must not confuse NextIndex.
	os.MkdirAll(filepath.Join(dir, "notes"), 0755)
	os.MkdirAll(filepath.Join(dir, DirName(1)), 0755)
	if n, err := NextIndex(dir); err != nil || n != 4 {
		t.Fatalf("NextIndex = %d, %v; want 4, nil", n, err)
	}
}
