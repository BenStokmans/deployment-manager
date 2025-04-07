package api

import (
	"encoding/json"
	"errors"
	"fmt"
	deployment_manager "github.com/BenStokmans/deployment-manager"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"net/http"
)

type DeploymentApi struct {
	config  deployment_manager.Config
	Router  *http.ServeMux
	Handler http.Handler
	Server  http.Server
}

func NewDeploymentApi(config deployment_manager.Config) *DeploymentApi {
	return &DeploymentApi{
		config: config,
	}
}

func (d *DeploymentApi) Start() error {
	// Start the API server
	// create http listener
	d.Router = http.NewServeMux()
	d.Router.HandleFunc("/deploy", d.HandleDeploy)
	d.Router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})
	d.Handler = c.Handler(d.Router)
	d.Server = http.Server{Addr: d.config.ApiUrl, Handler: d.Handler}

	if err := d.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logrus.Fatalf("failed to start server: %v", err)
	}
	logrus.Infof("server started on %s", d.config.ApiUrl)
	return nil
}

func (d *DeploymentApi) HandleDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	token := r.Header.Get("Authorization")
	if token != d.config.Token {
		w.WriteHeader(http.StatusUnauthorized)
	}

	name := r.URL.Query().Get("name")
	targetName := r.URL.Query().Get("target")
	result := d.DoDeploy(name, targetName)

	serialized, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(serialized)
	if err != nil {
		logrus.Errorf("failed to serialize deployment result: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (d *DeploymentApi) DoDeploy(name, targetName string) DeploymentResult {
	target, err := d.GetTarget(targetName)
	if err != nil {
		logrus.Errorf("target %s not found", targetName)
		return DeploymentResult{
			Status: UnknownTarget,
			Logs:   err.Error(),
		}
	}

	deployment, err := d.GetDeployment(name)
	if err != nil {
		logrus.Errorf("deployment %s not found", name)
		return DeploymentResult{
			Status: UnknownDeployment,
			Logs:   err.Error(),
		}
	}

	targetLogs, err := target.Execute()
	if err != nil {
		logrus.Errorf("target %s failed", target.Name)
		return DeploymentResult{
			Status: DeploymentStatusFailed,
			Logs:   targetLogs,
		}
	}
	logrus.Infof("target %s completed", target.Name)

	deploymentLogs, err := deployment.Execute()
	if err != nil {
		logrus.Errorf("deployment %s failed", deployment.Name)
		return DeploymentResult{
			Status: DeploymentStatusFailed,
			Logs:   deploymentLogs,
		}
	}

	logrus.Infof("deployment %s completed", deployment.Name)
	return DeploymentResult{
		Deployment: &deployment,
		Status:     DeploymentStatusCompleted,
		Logs:       targetLogs + "\n" + deploymentLogs,
	}
}

func (d *DeploymentApi) GetDeployment(name string) (deployment_manager.Deployment, error) {
	for _, deployment := range d.config.Deployments {
		if deployment.Name == name {
			return deployment, nil
		}
	}
	return deployment_manager.Deployment{}, fmt.Errorf("deployment %s not found", name)
}

func (d *DeploymentApi) GetTarget(name string) (deployment_manager.Target, error) {
	for _, target := range d.config.Targets {
		if target.Name == name {
			return target, nil
		}
	}
	return deployment_manager.Target{}, fmt.Errorf("target %s not found", name)
}

func (d *DeploymentApi) GetDeployments() ([]deployment_manager.Deployment, error) {
	deployments := make([]deployment_manager.Deployment, len(d.config.Deployments))
	for i, deployment := range d.config.Deployments {
		deployments[i] = deployment
	}
	return deployments, nil
}

func (d *DeploymentApi) GetTargets() ([]deployment_manager.Target, error) {
	targets := make([]deployment_manager.Target, len(d.config.Targets))
	for i, target := range d.config.Targets {
		targets[i] = target
	}
	return targets, nil
}
