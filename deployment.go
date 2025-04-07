package deployment_manager

import "os/exec"

type Deployment struct {
	// The name of the deployment
	Name string `yaml:"name"`
	// The commands to run for the deployment
	Commands []string `yaml:"commands"`
}

func (d *Deployment) Execute() (string, error) {
	// Execute the commands for the deployment
	completeOutput := ""
	for _, command := range d.Commands {
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return completeOutput, err
		}
		completeOutput += string(output)
	}
	return completeOutput, nil
}
