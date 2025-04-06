package main

import (
	"fmt"
	deployment_manager "github.com/BenStokmans/deployment-manager"
	"github.com/BenStokmans/deployment-manager/api"
	"gopkg.in/yaml.v3"
	"os"
)

const filePath = "/etc/deployment-manager/config.yaml"

func main() {
	var content []byte
	if _, err := os.ReadFile(filePath); err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return
	}

	var config deployment_manager.Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		fmt.Printf("Error unmarshalling YAML: %v\n", err)
		return
	}
	if config.ApiUrl == "" {
		fmt.Println("API URL is not set in the config file.")
		return
	}

	fmt.Printf("API URL: %s\n", config.ApiUrl)
	daemon := api.NewDeploymentApi(config)
	if err := daemon.Start(); err != nil {
		fmt.Printf("Error starting daemon: %v\n", err)
		return
	}
}
