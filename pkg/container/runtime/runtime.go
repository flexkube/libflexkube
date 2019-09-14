package runtime

import (
	"fmt"

	"github.com/invidian/etcd-ariadnes-thread/internal/util"
)

const defaultRuntime = "docker"

// runtimes is the map of registered container runtimes
var runtimes map[string]bool

func init() {
	runtimes = make(map[string]bool)
}

// Register should be used in each container implementation init() to register support
// for new container runtime
func Register(name string) {
	if _, exists := runtimes[name]; exists {
		panic(fmt.Sprintf("container runtime with name %q registered already", name))
	}
	runtimes[name] = true
}

// IsRuntimeSupported is a handy helper which unwraps default runtime name
// and then check if it is supported.
func IsRuntimeSupported(name string) bool {
	n := GetRuntimeName(name)
	_, exists := runtimes[n]
	return exists
}

// GetRuntimeName expands given runtime name
//
// If name is empty, it returns default runtime name
func GetRuntimeName(name string) string {
	return util.DefaultString(name, defaultRuntime)
}

// Runtime interface describes universal way of managing containers
// across different container runtimes.
type Runtime interface {
	// Create creates container and returns it's unique identifier
	Create(*Config) (string, error)
	// Delete removes the container
	Delete(string) error
	// Start starts created container
	Start(string) error
	// Status returns status of the container
	Status(string) (*Status, error)
	// Stop takes unique identifier as a parameter and stops the container
	Stop(string) error
}

// Config describes how container should be created
type Config struct {
	Name       string   `json:"name" yaml:"name"`
	Image      string   `json:"image" yaml:"image"`
	Args       []string `json:"args,omitempty" yaml:"args,omitempty"`
	Entrypoint []string `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
}

// Status describes what informations are returned about container
type Status struct {
	Image  string `json:"image" yaml:"image"`
	ID     string `json:"id" yaml:"id"`
	Name   string `json:"name", yaml:"name"`
	Status string `json:"status", yaml:"status"`
}
