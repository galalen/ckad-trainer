package checks

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/galalen/ckad-trainer/models"
)

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		in      string
		want    []string
		wantErr bool
	}{
		{in: "kubectl get pods", want: []string{"kubectl", "get", "pods"}},
		{in: `kubectl get deploy web -o jsonpath='{.spec.replicas}'`, want: []string{"kubectl", "get", "deploy", "web", "-o", "jsonpath={.spec.replicas}"}},
		{in: `kubectl exec p -- sh -c "wget -qO- web | head -1"`, want: []string{"kubectl", "exec", "p", "--", "sh", "-c", "wget -qO- web | head -1"}},
		{in: `kubectl get pods -o jsonpath="{.items[?(@.status.phase==\"Running\")].metadata.name}"`, want: []string{"kubectl", "get", "pods", "-o", `jsonpath={.items[?(@.status.phase=="Running")].metadata.name}`}},
		{in: "kubectl get pods | grep web", wantErr: true},
		{in: "kubectl get pods; rm -rf /", wantErr: true},
		{in: "kubectl get pods $(id)", wantErr: true},
		{in: "kubectl get 'pods", wantErr: true},
	}

	for _, tt := range tests {
		got, err := SplitArgs(tt.in)
		if (err != nil) != tt.wantErr {
			t.Fatalf("SplitArgs(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
		}
		if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
			t.Errorf("SplitArgs(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseCommand(t *testing.T) {
	allowed := []string{
		"kubectl get pods -n ns",
		"kubectl -n ns get pods",
		"kubectl --namespace ns describe svc web",
		"kubectl auth can-i get pods --as system:serviceaccount:ns:sa",
		"kubectl rollout status deployment/web -n ns",
		"kubectl -n ns rollout history deploy web",
		"kubectl logs job/pi -n ns",
	}
	for _, c := range allowed {
		if _, err := ParseCommand(c); err != nil {
			t.Errorf("ParseCommand(%q) unexpected error: %v", c, err)
		}
	}

	rejected := []string{
		"kubectl delete pod web",
		"kubectl -n ns apply -f x.yaml",
		"kubectl rollout undo deployment/web",
		"kubectl rollout restart deployment/web",
		"kubectl auth reconcile -f rbac.yaml",
		"curl http://example.com",
		"bash -c 'kubectl get pods'",
		"",
	}
	for _, c := range rejected {
		if _, err := ParseCommand(c); err == nil {
			t.Errorf("ParseCommand(%q) expected error", c)
		}
	}
}

func fakeRunner(out string, err error) Runner {
	return func(ctx context.Context, args []string) (string, error) {
		return out, err
	}
}

func TestRunCheck(t *testing.T) {
	cmd := "kubectl get deploy web -o jsonpath='{.spec.replicas}'"
	tests := []struct {
		name   string
		check  models.Check
		out    string
		runErr error
		want   bool
	}{
		{name: "equals", check: models.Check{Cmd: cmd, Expect: "3"}, out: "3\n", want: true},
		{name: "equals quoted", check: models.Check{Cmd: cmd, Expect: "'3'"}, out: "3", want: true},
		{name: "equals mismatch", check: models.Check{Cmd: cmd, Expect: "3"}, out: "2", want: false},
		{name: "contains", check: models.Check{Cmd: cmd, Expect: "nginx", Match: MatchContains}, out: "image nginx:1.25", want: true},
		{name: "regex", check: models.Check{Cmd: cmd, Expect: `^nginx:1\.\d+$`, Match: MatchRegex}, out: "nginx:1.25", want: true},
		{name: "exit0 default", check: models.Check{Cmd: cmd}, out: "anything", want: true},
		{name: "command fails", check: models.Check{Cmd: cmd}, runErr: errors.New("NotFound"), want: false},
		{name: "disallowed", check: models.Check{Cmd: "kubectl delete pod web"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RunCheck(context.Background(), fakeRunner(tt.out, tt.runErr), tt.check, time.Second)
			if r.Passed != tt.want {
				t.Errorf("Passed = %v, want %v (err: %v)", r.Passed, tt.want, r.Err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	good := models.Scenario{Tasks: []models.Task{{Checks: []models.Check{{Cmd: "kubectl get pods"}}}}}
	if err := Validate(good); err != nil {
		t.Fatalf("Validate(good) = %v", err)
	}

	bad := []models.Scenario{
		{},
		{Tasks: []models.Task{{Text: "no checks"}}},
		{Tasks: []models.Task{{Checks: []models.Check{{Cmd: "kubectl delete ns x"}}}}},
		{Tasks: []models.Task{{Checks: []models.Check{{Cmd: "kubectl get pods", Expect: "(", Match: MatchRegex}}}}},
	}
	for i, s := range bad {
		if err := Validate(s); err == nil {
			t.Errorf("Validate(bad[%d]) expected error", i)
		}
	}
}
