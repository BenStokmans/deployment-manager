package main

import (
	"flag"
	"fmt"
	deployment_manager "github.com/BenStokmans/deployment-manager"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
)

var fileFlag = flag.String("file", "", "Path to the config file")

const repoLink = "https://github.com/BenStokmans/deployment-manager"

func main() {
	flag.Parse()

	if *fileFlag == "" {
		fmt.Println("Please provide a config file using the -file flag")
		return
	}

	var content []byte
	if _, err := os.ReadFile(*fileFlag); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	var config deployment_manager.Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		fmt.Printf("Error unmarshalling YAML: %v\n", err)
		return
	}

	CheckRoot()

	InstallLatestDaemon(&config)
}

func CheckRoot() {
	if os.Geteuid() != 0 {
		fmt.Println("This program must be run as root.")
		os.Exit(1)
	}
}

const serviceString = `[Unit]
Description=Deployment Manager Daemon
After=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/deployment-manager-daemon
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

# Security options
ProtectSystem=full
PrivateTmp=true
NoNewPrivileges=false  # Allow privileged operations since we're running as root

[Install]
WantedBy=multi-user.target`

func InstallLatestDaemon(config *deployment_manager.Config) {
	// Add logic to build the latest daemon here
	fmt.Println("Building the latest daemon...")

	originalDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	// Get user cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		fmt.Printf("Error getting user cache directory: %v\n", err)
		return
	}

	tmpDir := cacheDir + "/deployment-manager"

	// remove existing directory
	if _, err := os.Stat(tmpDir); err == nil {
		cmd := exec.Command("rm", "-rf", tmpDir)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error removing existing directory: %v\n", err)
			return
		}
	}

	// Create the directory
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	// Clone the repository
	cmd := exec.Command("git", "clone", repoLink)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error cloning repository: %s\n", string(output))
		return
	}

	err = os.Chdir(tmpDir + "/deployment-manager/cmd/daemon")
	if err != nil {
		fmt.Printf("Error changing directory: %v\n", err)
		return
	}

	// Build the daemon
	cmd = exec.Command("go", "build", "-o", "daemon")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error building daemon: %v\n", err)
		return
	}

	// remove if daemon already exists
	if _, err := os.Stat("/usr/local/bin/deployment-manager-daemon"); err == nil {
		cmd = exec.Command("rm", "/usr/local/bin/deployment-manager-daemon")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error removing existing daemon: %v\n", err)
			return
		}
	}

	// Move the daemon to /usr/local/bin
	cmd = exec.Command("mv", "daemon", "/usr/local/bin/deployment-manager-daemon")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error moving daemon: %v\n", err)
		return
	}

	// Clean up the temporary directory
	cmd = exec.Command("rm", "-rf", tmpDir)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error removing temporary directory: %v\n", err)
		return
	}

	// Set permissions
	cmd = exec.Command("chmod", "755", "/usr/local/bin/deployment-manager-daemon")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error setting permissions: %v\n", err)
		return
	}

	// install config file
	configDir := "/etc/deployment-manager"
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Error creating config directory: %v\n", err)
		return
	}

	// remove existing config file
	if _, err := os.Stat(configDir + "/config.yaml"); err == nil {
		cmd = exec.Command("rm", configDir+"/config.yaml")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error removing existing config file: %v\n", err)
			return
		}
	}

	configFile := configDir + "/config.yaml"
	config.Token, err = CreateToken()
	if err != nil {
		fmt.Printf("Error creating token: %v\n", err)
		return
	}

	content, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("Error marshalling config to YAML: %v\n", err)
		return
	}

	// log token for user
	fmt.Printf("Token: %s\n", config.Token)

	// Write the config to a file
	if err := os.WriteFile(configFile, content, 0644); err != nil {
		fmt.Printf("Error writing config file: %v\n", err)
		return
	}

	// remove existing service
	if _, err := os.Stat("/etc/systemd/system/deployment-manager-daemon.service"); err == nil {
		// disable and stop the service
		cmd = exec.Command("systemctl", "stop", "deployment-manager-daemon")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error stopping systemd service: %v\n", err)
			return
		}

		cmd = exec.Command("systemctl", "disable", "deployment-manager-daemon")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error disabling systemd service: %v\n", err)
			return
		}

		cmd = exec.Command("rm", "/etc/systemd/system/deployment-manager-daemon.service")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error removing systemd service: %v\n", err)
			return
		}
	}

	// install systemctl service
	cmd = exec.Command("bash", "-c", fmt.Sprintf("echo '%s' > /etc/systemd/system/deployment-manager-daemon.service", serviceString))
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error installing systemd service: %v\n", err)
		return
	}

	cmd = exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error reloading systemd: %v\n", err)
		return
	}

	cmd = exec.Command("systemctl", "enable", "deployment-manager-daemon")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error enabling systemd service: %v\n", err)
		return
	}

	cmd = exec.Command("systemctl", "start", "deployment-manager-daemon")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error starting systemd service: %v\n", err)
		return
	}

	fmt.Println("Daemon installed successfully.")
	// change back to original directory
	if err := os.Chdir(originalDir); err != nil {
		fmt.Printf("Error changing back to original directory: %v\n", err)
		return
	}
}

func CreateToken() (string, error) {
	// Add logic to create a token here
	fmt.Println("Creating a token...")
	// Check if the user is root
	if os.Geteuid() != 0 {
		fmt.Println("This program must be run as root.")
		os.Exit(1)
	}

	// Generate a random token
	cmd := exec.Command("openssl", "rand", "-base64", "32")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		return "", err
	}

	token := string(output)
	// Remove trailing newline
	token = token[:len(token)-1]

	return token, nil
}
