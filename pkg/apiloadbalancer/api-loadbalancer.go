package apiloadbalancer

// TODO figure out better name for this package, maybe something more generic?
// ^ This comment is below the package keyword, to prevent golint from complaining

import (
	"fmt"
	"strings"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/container/runtime/docker"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
	"github.com/flexkube/libflexkube/pkg/host"
)

// APILoadBalancer is a user-configurable representation of single instance of API load balancer
type APILoadBalancer struct {
	Image              string     `json:"image,omitempty" yaml:"image,omitempty"`
	Host               *host.Host `json:"host,omitempty" yaml:"host,omitempty"`
	MetricsBindAddress string     `json:"metricsBindAddress,omitempty" yaml:"metricsBindAddress,omitempty"`
	MetricsBindPort    int        `json:"metricsBindPort,omitempty" yaml:"metricsBindPort,omitempty"`
	Servers            []string   `json:"servers,omitempty" yaml:"servers,omitempty"`
}

// apiLoadBalancer is validated and executable version of APILoadBalancer
type apiLoadBalancer struct {
	image              string
	host               *host.Host
	servers            []string
	metricsBindAddress string
	metricsBindPort    int
}

// ToHostConfiguredContainer takes configuration stored in the struct and converts it to HostConfiguredContainer
// which can be then added to Containers struct and executed
//
// TODO ToHostConfiguredContainer should become an interface, since we use this pattern in all packages
func (a *apiLoadBalancer) ToHostConfiguredContainer() *container.HostConfiguredContainer {
	servers := []string{}
	for i, s := range a.servers {
		servers = append(servers, fmt.Sprintf("server %d %s:8443 check", i, s))
	}

	configFiles := make(map[string]string)
	configFiles["/etc/haproxy/haproxy.cfg"] = fmt.Sprintf(`defaults
	# Do TLS passthrough
	mode tcp
	# Required values for both frontend and backend
	timeout connect 5000ms
	timeout client 50000ms
	timeout server 50000ms

frontend kube-apiserver
	# TODO make it configurable
	bind 0.0.0.0:6443
	default_backend kube-apiserver

backend kube-apiserver
	%s

frontend stats
	bind 0.0.0.0:%d
	mode http
	option http-use-htx
	http-request use-service prometheus-exporter if { path /metrics }
	stats enable
	stats uri /stats
	stats refresh 10s
`, strings.Join(servers, "\n	"), a.metricsBindPort)

	c := container.Container{
		// TODO this is weird. This sets docker as default runtime config
		Runtime: container.RuntimeConfig{
			Docker: &docker.Config{},
		},
		Config: types.ContainerConfig{
			// TODO make it configurable? And don't force user to use HAProxy
			Name:  "api-loadbalancer-haproxy",
			Image: a.image,
			Ports: []types.PortMap{
				{
					Protocol: "tcp",
					Port:     6443,
				},
				{
					Protocol: "tcp",
					Port:     a.metricsBindPort,
					IP:       a.metricsBindAddress,
				},
			},
			Mounts: []types.Mount{
				{
					Source: "/etc/haproxy/haproxy.cfg",
					Target: "/usr/local/etc/haproxy/haproxy.cfg",
				},
			},
		},
	}

	return &container.HostConfiguredContainer{
		Host:        *a.host,
		ConfigFiles: configFiles,
		Container:   c,
	}
}

// New validates APILoadBalancer configuration and fills it with default options
// If configuration is wrong, error is returned.
//
// TODO I think we shouldn't fill the default values here. Maybe do it one level up?
func (a *APILoadBalancer) New() (*apiLoadBalancer, error) {
	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate API Load balancer configuration: %w", err)
	}

	na := &apiLoadBalancer{
		image:              a.Image,
		host:               a.Host,
		servers:            a.Servers,
		metricsBindAddress: a.MetricsBindAddress,
		metricsBindPort:    a.MetricsBindPort,
	}

	// Fill empty fields with default values
	if na.image == "" {
		na.image = defaults.HAProxyImage
	}

	if na.metricsBindPort == 0 {
		na.metricsBindPort = 8080
	}

	return na, nil
}

// Validate contains all validation rules for APILoadBalancer struct.
// This method can be used by the user to catch configuration errors early.
func (a *APILoadBalancer) Validate() error {
	if a.Host == nil {
		return fmt.Errorf("field Host must be set")
	}

	if len(a.Servers) == 0 {
		return fmt.Errorf("at least one server must be set")
	}

	if a.MetricsBindAddress == "" {
		return fmt.Errorf("field MetricsBindAddress must be set")
	}

	return nil
}
