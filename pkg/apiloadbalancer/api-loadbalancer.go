package apiloadbalancer

// TODO figure out better name for this package, maybe something more generic?
// ^ This comment is below the package keyword, to prevent golint from complaining

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
)

// APILoadBalancer is a user-configurable representation of single instance of API load balancer
type APILoadBalancer struct {
	Image          string    `json:"image,omitempty"`
	Host           host.Host `json:"host,omitempty"`
	Servers        []string  `json:"servers,omitempty"`
	Name           string    `json:"name,omitempty"`
	HostConfigPath string    `json:"hostConfigPath,omitempty"`
	BindAddress    string    `json:"bindAddress,omitempty"`
}

// apiLoadBalancer is validated and executable version of APILoadBalancer
type apiLoadBalancer struct {
	image          string
	host           host.Host
	servers        []string
	name           string
	hostConfigPath string
	bindAddress    string
}

func (a apiLoadBalancer) config() (string, error) {
	c := `
defaults
  # Do TLS passthrough
  mode tcp
  # Required values for both frontend and backend
  timeout connect 5000ms
  timeout client 50000ms
  timeout server 50000ms

frontend kube-apiserver
  bind {{ .BindAddress }}
  default_backend kube-apiserver

backend kube-apiserver
  {{- range $i, $s := .Servers }}
  server {{ $i }} {{ $s }} check
  {{- end }}
`

	t := template.Must(template.New("haproxy.cfg").Parse(c))

	var buf bytes.Buffer

	d := struct {
		Servers     []string
		BindAddress string
	}{
		a.servers,
		a.bindAddress,
	}

	if err := t.Execute(&buf, d); err != nil {
		return "", fmt.Errorf("executing template failed: %w", err)
	}

	return fmt.Sprintf("%s\n", strings.TrimSpace(buf.String())), nil
}

const (
	hostConfigPath      = "/etc/haproxy/haproxy.cfg"
	containerConfigPath = "/usr/local/etc/haproxy/haproxy.cfg"
	containerName       = "api-loadbalancer-haproxy"
)

// ToHostConfiguredContainer takes configuration stored in the struct and converts it to HostConfiguredContainer
// which can be then added to Containers struct and executed
//
// TODO ToHostConfiguredContainer should become an interface, since we use this pattern in all packages
func (a *apiLoadBalancer) ToHostConfiguredContainer() (*container.HostConfiguredContainer, error) {
	config, err := a.config()
	if err != nil {
		return nil, fmt.Errorf("failed generating config: %w", err)
	}

	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: docker.DefaultConfig(),
		},
		Config: types.ContainerConfig{
			// TODO make it configurable? And don't force user to use HAProxy
			Name:        a.name,
			Image:       a.image,
			NetworkMode: "host",
			// Run as unprivileged user.
			User: "65534",
			Mounts: []types.Mount{
				{
					Source: a.hostConfigPath,
					Target: containerConfigPath,
				},
			},
		},
	}

	return &container.HostConfiguredContainer{
		Host: a.host,
		ConfigFiles: map[string]string{
			a.hostConfigPath: config,
		},
		Container: c,
	}, nil
}

// New validates APILoadBalancer configuration and fills it with default options
// If configuration is wrong, error is returned.
//
// TODO I think we shouldn't fill the default values here. Maybe do it one level up?
func (a *APILoadBalancer) New() (container.ResourceInstance, error) {
	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate API Load balancer configuration: %w", err)
	}

	na := &apiLoadBalancer{
		image:          a.Image,
		host:           a.Host,
		servers:        a.Servers,
		name:           util.PickString(a.Name, containerName),
		hostConfigPath: util.PickString(a.HostConfigPath, hostConfigPath),
		bindAddress:    a.BindAddress,
	}

	// Fill empty fields with default values
	if na.image == "" {
		na.image = defaults.HAProxyImage
	}

	return na, nil
}

// Validate contains all validation rules for APILoadBalancer struct.
// This method can be used by the user to catch configuration errors early.
func (a *APILoadBalancer) Validate() error {
	if len(a.Servers) == 0 {
		return fmt.Errorf("at least one server must be set")
	}

	if a.BindAddress == "" {
		return fmt.Errorf("bindAddress can't be empty")
	}

	return nil
}
