package deployment_manager

type Config struct {
	Token       string       `json:"token" yaml:"token,omitempty"`
	ApiUrl      string       `json:"api_url" yaml:"api_url"`
	Deployments []Deployment `json:"deployments" yaml:"deployments"`
	Targets     []Target     `json:"targets" yaml:"targets"`
}
