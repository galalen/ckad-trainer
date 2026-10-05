package llm

import (
	"fmt"
	"strings"
)

const systemPrompt = "You are a Kubernetes expert specializing in CKAD exam preparation. You reply with a single JSON object and nothing else."

// Domains are the CKAD skill areas scenarios rotate through.
var Domains = []string{
	"Pod creation and configuration",
	"ConfigMaps and Secrets",
	"Multi-container Pod patterns",
	"Deployments and rolling updates",
	"Services and networking",
	"Volumes and persistence",
	"Jobs and CronJobs",
	"Ingress resources",
	"Pod design with labels and selectors",
	"Horizontal Pod Autoscaler",
	"Probes and observability",
	"SecurityContext and ServiceAccounts",
	"Resource requests, limits and quotas",
	"NetworkPolicies",
}

// BuildPrompt asks for a single scenario in the given domain. avoid lists
// titles already generated in this run so parallel requests stay distinct.
func BuildPrompt(domain string, avoid []string) string {
	avoidText := ""
	if len(avoid) > 0 {
		avoidText = "\nDo not repeat these existing scenarios: " + strings.Join(avoid, "; ") + ".\n"
	}

	return fmt.Sprintf(`Generate ONE CKAD (Certified Kubernetes Application Developer) practice scenario.

Skill area: %s
%s
Requirements:
- Use only public images with exact tags (e.g. nginx:1.25, busybox:1.36, redis:7.2).
- It must run on any local cluster (minikube, kind, Docker Desktop).
- All resources live in the namespace written literally as {{ns}}. Use {{ns}} everywhere a namespace is needed (task text, setup, cleanup, solutions, checks). Never hard-code a namespace.
- 3 to 5 tasks. Each task is one concrete, exam-style instruction that names every resource exactly.
- Each task has a short hint (nudge, not the answer) and a solution (the kubectl commands or YAML that solve it).
- Each task has 1 to 3 checks that verify it.

Check rules (checks are executed WITHOUT a shell):
- "cmd" must be a single read-only kubectl command: get, describe, logs, auth can-i, rollout status, rollout history, top or exec.
- No pipes, redirects, &&, ;, $( ), or backticks.
- Prefer "kubectl get <kind> <name> -n {{ns}} -o jsonpath='{...}'" so the output is a single exact value.
- "match" is one of: "equals" (trimmed output equals "expect"), "contains", "regex", "exit0" (command succeeds; "expect" is empty).
- "expect" is exactly what the command prints when the task is solved correctly.

"setup" lists shell commands that prepare the cluster before the user starts. Always start with "kubectl create namespace {{ns}}" and add any pre-existing resources the tasks rely on.
"cleanup" lists commands that remove everything, normally "kubectl delete namespace {{ns}}".

Reply with this JSON object only:
{
  "title": "Short scenario title",
  "description": "Background and context for the scenario",
  "domain": "%s",
  "difficulty": "easy | medium | hard",
  "setup": ["kubectl create namespace {{ns}}"],
  "cleanup": ["kubectl delete namespace {{ns}}"],
  "tasks": [
    {
      "text": "Create a Deployment named web in namespace {{ns}} with 3 replicas of nginx:1.25",
      "hint": "kubectl create deployment supports --replicas",
      "solution": "kubectl create deployment web --image=nginx:1.25 --replicas=3 -n {{ns}}",
      "checks": [
        {
          "description": "Deployment web has 3 replicas",
          "cmd": "kubectl get deployment web -n {{ns}} -o jsonpath='{.spec.replicas}'",
          "expect": "3",
          "match": "equals"
        }
      ]
    }
  ]
}`, domain, avoidText, domain)
}
