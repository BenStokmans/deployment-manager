package deployment_manager

type Config struct {
	Token       string       `yaml:"token,omitempty"`
	ApiUrl      string       `yaml:"api_url"`
	Deployments []Deployment `yaml:"deployments"`
	Targets     []Target     `yaml:"targets"`
}
