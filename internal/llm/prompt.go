package llm

import (
	"fmt"
)

const systemPrompt = "You are a Kubernetes expert specializing in CKAD exam preparation."

func BuildPrompt(scenariosCount int) string {
	return fmt.Sprintf(`
	Generate %d detailed CKAD (Certified Kubernetes Application Developer) exam practice scenarios with the following enhanced requirements:
	Scenario Generation Guidelines
	
	Local Reproducibility
	
	Use only publicly available container images from Docker Hub or official Kubernetes repositories
	Specify exact image tags and versions
	Ensure scenarios can be run on:
	
	Minikube
	kind (Kubernetes in Docker)
	Docker Desktop Kubernetes
	Any standard local Kubernetes cluster

	Scenario Structure
	Each scenario should include:
	
	Title
	Description with clear context
	Specific, actionable tasks
	Validation script (returns non-zero exit code on failure)
	Recommended cluster setup commands
	Cleanup instructions
	
	
	Skill Areas to Cover
	
	Pod creation and configuration
	ConfigMaps and Secrets
	Multi-container Pod patterns
	Deployments and rolling updates
	Services and networking
	Volumes and persistence
	Jobs and CronJobs
	Ingress resources
	Pod design with labels and selectors
	Horizontal Pod Autoscaler

	Format the response as structured JSON with the following schema:
	{
	  "scenarios": [
		{
		  "title": "Scenario title",
		  "description": "Background and context",
		  "tasks": ["Task 1", "Task 2", "Task 3", ...],
		  "validation": "#!/bin/bash\n# Validation script that checks the solution"
		}
	  ]
	}`,
		scenariosCount,
	)
}
