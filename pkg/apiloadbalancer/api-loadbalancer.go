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

// APILoadBalancer is a user-configurable representation of single instance of API load balancer.
type APILoadBalancer struct {
	// Image allows to set Docker image with tag, which will be used by the container.
	// if instance itself has no image set. If empty, haproxy image defined in pkg/defaults
	// will be used.
	//
	// Example value: 'haproxy:2.1.4-alpine'
	//
	// This field is optional.
	Image string `json:"image,omitempty"`

	// Host describes on which machine member container should be created.
	//
	// This field is required.
	Host host.Host `json:"host,omitempty"`

	// Servers is a list of Kubernetes API server addresses, which should be used as a backend
	// servers.
	//
	// Example value: '[]string{"192.168.10.10:6443", "192.168.10.11:6443"}'.
	//
	// This field is optional, if used together with APILoadBalancers struct.
	Servers []string `json:"servers,omitempty"`

	// Name is a container name to create. If you want to run more than one instance on a
	// single host, this field must be unique, otherwise you will get an error with duplicated
	// container name.
	//
	// Currently, when you want to expose Kubernetes API using the load balancer on more than
	// one specific address (so not listening on 0.0.0.0), then two pools are required.
	// This limitation will be addressed in the future.
	//
	// This field is optional. If empty, value from ContainerName constant will be used.
	Name string `json:"name,omitempty"`

	// HostConfigPath is a path on the host filesystem, where load balancer configuration should be
	// written. If you want to run more than one instance on a single host, this field must
	// be unique, otherwise the configuration will be overwritten by the other instance.
	//
	// Currently, when you want to expose Kubernetes API using the load balancer on more than
	// one specific address (so not listening on 0.0.0.0), then two pools are required.
	// This limitation will be addressed in the future.
	//
	// This field is optional. If empty, value from HostConfigPath constant will be used.
	HostConfigPath string `json:"hostConfigPath,omitempty"`

	// BindAddress controls, on which IP address load balancer container will be listening on.
	// Usually, it is set to here to either 127.0.0.1 to only expose the load balancer on host's
	// localhost address or to 0.0.0.0 to expose the API on all interfaces.
	//
	// If you want to listen on specific address, which is different for each host, set it for each
	// LoadBalancer instance.
	//
	// This field is optional, if used together with APILoadBalancers struct.
	BindAddress string `json:"bindAddress,omitempty"`
}

// apiLoadBalancer is validated and executable version of APILoadBalancer.
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
  timeout connect 5s
  timeout client 30s
  timeout client-fin 30s
  timeout server 30s
  timeout tunnel 21d

frontend kube-apiserver
  bind {{ .BindAddress }}
  default_backend kube-apiserver

backend kube-apiserver
  option httpchk GET /healthz HTTP/1.1\r\nHost:\ kube-apiserver
  {{- range $i, $s := .Servers }}
  server {{ $i }} {{ $s }} verify none check check-ssl
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
	// HostConfigPath is a default path on the host filesystem, where container
	// configuration will be stored.
	HostConfigPath = "/etc/haproxy/haproxy.cfg"

	// ContainerName is a default name for load balancer container.
	ContainerName = "api-loadbalancer-haproxy"

	// Path inside the container, where configuration
	// stored on the host filesystem should be mapped into.
	containerConfigPath = "/usr/local/etc/haproxy/haproxy.cfg"
)

// ToHostConfiguredContainer takes configuration stored in the struct and converts it to HostConfiguredContainer
// which can be then added to Containers struct and executed.
//
// TODO: ToHostConfiguredContainer should become an interface, since we use this pattern in all packages.
func (a *apiLoadBalancer) ToHostConfiguredContainer() (*container.HostConfiguredContainer, error) {
	config, err := a.config()
	if err != nil {
		return nil, fmt.Errorf("generating config: %w", err)
	}

	c := container.Container{
		// TODO: This is weird. This sets docker as default runtime config.
		Runtime: container.RuntimeConfig{
			Docker: docker.DefaultConfig(),
		},
		Config: types.ContainerConfig{
			// TODO: Make it configurable? And don't force user to use HAProxy.
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
// TODO: I think we shouldn't fill the default values here. Maybe do it one level up?
func (a *APILoadBalancer) New() (container.ResourceInstance, error) {
	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("validating API Load balancer configuration: %w", err)
	}

	na := &apiLoadBalancer{
		image:          a.Image,
		host:           a.Host,
		servers:        a.Servers,
		name:           util.PickString(a.Name, ContainerName),
		hostConfigPath: util.PickString(a.HostConfigPath, HostConfigPath),
		bindAddress:    a.BindAddress,
	}

	// Fill empty fields with default values.
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
