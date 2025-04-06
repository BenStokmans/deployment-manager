package deployment_manager

import "os/exec"

type Target struct {
	// The name of the target
	Name string `json:"name" yaml:"name"`
	// The commands to run for the target
	Commands []string `json:"commands" yaml:"commands"`
}

func (t *Target) Execute() (string, error) {
	// Execute the commands for the deployment
	completeOutput := ""
	for _, command := range t.Commands {
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return completeOutput, err
		}
		completeOutput += string(output)
	}
	return completeOutput, nil
}
