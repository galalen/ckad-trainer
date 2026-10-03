package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/galalen/ckad-trainer/config"
)

const validScenario = `{"title":"Web","description":"d","tasks":[{"text":"t","checks":[{"cmd":"kubectl get deploy web -n {{ns}}"}]}]}`

func TestExtractJSON(t *testing.T) {
	tests := map[string]string{
		"```json\n{\"a\":1}\n```":          `{"a":1}`,
		`Here you go: {"a":"}"} trailing}`: `{"a":"}"}`,
		`{"a":{"b":"\"{"}}`:                `{"a":{"b":"\"{"}}`,
		"no json":                          "",
		`{"unterminated":`:                 "",
	}
	for in, want := range tests {
		if got := extractJSON(in); got != want {
			t.Errorf("extractJSON(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseScenarioRejectsUnsafeChecks(t *testing.T) {
	if _, err := parseScenario(validScenario); err != nil {
		t.Fatalf("parseScenario(valid) = %v", err)
	}
	unsafe := strings.Replace(validScenario, "kubectl get", "kubectl delete", 1)
	if _, err := parseScenario(unsafe); err == nil {
		t.Fatal("expected unsafe check to be rejected")
	}
}

// rewrite sends every request to the test server regardless of host.
type rewrite struct{ target string }

func (r rewrite) RoundTrip(req *http.Request) (*http.Response, error) {
	u := *req.URL
	u.Scheme, u.Host = "http", r.target
	req = req.Clone(req.Context())
	req.URL = &u
	return http.DefaultTransport.RoundTrip(req)
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	cfg := config.NewConfig()
	cfg.APIKey = "test"
	c := NewClient(cfg)
	c.http.Transport = rewrite{target: strings.TrimPrefix(srv.URL, "http://")}
	c.backoff = time.Millisecond
	return c
}

func reply(w http.ResponseWriter, content string) {
	json.NewEncoder(w).Encode(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"content": content}}},
	})
}

func TestGenerateScenarioRetries(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch calls.Add(1) {
		case 1:
			http.Error(w, "rate limited", http.StatusTooManyRequests)
		case 2:
			reply(w, "not json at all")
		default:
			reply(w, validScenario)
		}
	})

	s, err := c.GenerateScenario(context.Background(), Domains[0], nil)
	if err != nil {
		t.Fatalf("GenerateScenario = %v", err)
	}
	if s.Title != "Web" || calls.Load() != 3 {
		t.Errorf("got title %q after %d calls", s.Title, calls.Load())
	}
}

func TestGenerateScenarioDoesNotRetryAuthErrors(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "bad key", http.StatusUnauthorized)
	})

	_, err := c.GenerateScenario(context.Background(), Domains[0], nil)
	if err == nil || calls.Load() != 1 {
		t.Fatalf("err = %v after %d calls; want one failed call", err, calls.Load())
	}
}

func TestRequestUsesProviderDefaultModel(t *testing.T) {
	var model string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model string `json:"model"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		model = body.Model
		reply(w, validScenario)
	})
	c.config.Provider = config.OpenAI

	if _, err := c.GenerateScenario(context.Background(), Domains[0], nil); err != nil {
		t.Fatal(err)
	}
	if model != c.config.GetDefaultModel() || strings.Contains(model, "deepseek") {
		t.Errorf("openai request used model %q", model)
	}
}
