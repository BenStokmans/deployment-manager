package api

import deployment_manager "github.com/BenStokmans/deployment-manager"

// DeploymentResult represents the result of a deployment.
type DeploymentResult struct {
	deployment_manager.Deployment
	// Status is the status of the deployment.
	Status string `json:"status" yaml:"status"`
	// Logs is the logs of the deployment.
	Logs string `json:"logs" yaml:"logs"`
}
