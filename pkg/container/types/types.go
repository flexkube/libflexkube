package types

// To avoid cyclic dependencies between container, runtime and runtime implementation packages,
// we put container-related types in separated package.

// ContainerConfig stores runtime-agnostic information how to run the container
type ContainerConfig struct {
	Name       string   `json:"name" yaml:"name"`
	Image      string   `json:"image" yaml:"image"`
	Args       []string `json:"args,omitempty" yaml:"args,omitempty"`
	Entrypoint []string `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
}

// ContainerStatus stores status information received from the runtime
type ContainerStatus struct {
	Image  string `json:"image" yaml:"image"`
	ID     string `json:"id" yaml:"id"`
	Name   string `json:"name", yaml:"name"`
	Status string `json:"status", yaml:"status"`
}
