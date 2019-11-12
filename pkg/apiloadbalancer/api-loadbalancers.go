// TODO figure out better name for this package, maybe something more generic?
package apiloadbalancer

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
)

type APILoadBalancers struct {
	Image            string            `json:"image,omitempty" yaml:"image,omitempty"`
	SSH              *ssh.Config       `json:"ssh,omitempty" yaml:"ssh,omitempty"`
	Servers          []string          `json:"servers,omitempty" yaml:"servers,omitempty"`
	APILoadBalancers []APILoadBalancer `json:"apiLoadBalancers,omitempty" yaml:"apiLoadBalancers,omitempty"`

	// Serializable fields
	State container.ContainersState `json:"state:omitempty" yaml:"state,omitempty"`
}

type apiLoadBalancers struct {
	image      string
	ssh        *ssh.Config
	containers container.Containers
}

func (a *APILoadBalancers) New() (*apiLoadBalancers, error) {
	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate API Load balancers configuration: %w", err)
	}

	apiLoadBalancers := &apiLoadBalancers{
		image: a.Image,
		ssh:   a.SSH,
		containers: container.Containers{
			PreviousState: a.State,
			DesiredState:  make(container.ContainersState),
		},
	}

	for i, lb := range a.APILoadBalancers {
		if lb.Image == "" && a.Image != "" {
			lb.Image = a.Image
		}
		if len(lb.Servers) <= 0 && len(a.Servers) > 0 {
			lb.Servers = a.Servers
		}

		// TODO find better way to handle defaults!!!
		/*if lb.Host == nil || (lb.Host.DirectConfig == nil && lb.Host.SSHConfig == nil) {
			lb.Host = &host.Host{
				DirectConfig: &direct.DirectConfig{},
			}
		}*/
		if lb.Host == nil {
			lb.Host = &host.Host{
				SSHConfig: a.SSH,
			}
		}
		if lb.Host != nil && lb.Host.SSHConfig != nil && lb.Host.SSHConfig.PrivateKey == "" && a.SSH != nil && a.SSH.PrivateKey != "" {
			lb.Host.SSHConfig.PrivateKey = a.SSH.PrivateKey
		}

		if lb.Host != nil && lb.Host.SSHConfig != nil && lb.Host.SSHConfig.User == "" && a.SSH != nil && a.SSH.User != "" {
			lb.Host.SSHConfig.User = a.SSH.User
		}
		if lb.Host != nil && lb.Host.SSHConfig != nil && lb.Host.SSHConfig.User == "" {
			lb.Host.SSHConfig.User = "root"
		}

		if lb.Host != nil && lb.Host.SSHConfig != nil && lb.Host.SSHConfig.ConnectionTimeout == "" && a.SSH != nil && a.SSH.ConnectionTimeout != "" {
			lb.Host.SSHConfig.ConnectionTimeout = a.SSH.ConnectionTimeout
		}
		if lb.Host != nil && lb.Host.SSHConfig != nil && lb.Host.SSHConfig.ConnectionTimeout == "" {
			lb.Host.SSHConfig.ConnectionTimeout = "30s"
		}

		if lb.Host != nil && lb.Host.SSHConfig != nil && lb.Host.SSHConfig.Port == 0 && a.SSH != nil && a.SSH.Port != 0 {
			lb.Host.SSHConfig.Port = a.SSH.Port
		}
		if lb.Host != nil && lb.Host.SSHConfig != nil && lb.Host.SSHConfig.Port == 0 {
			lb.Host.SSHConfig.Port = 22
		}

		lbx, err := lb.New()
		if err != nil {
			return nil, fmt.Errorf("that was unexpected: %w", err)
		}
		apiLoadBalancers.containers.DesiredState[strconv.Itoa(i)] = lbx.ToHostConfiguredContainer()
	}

	return apiLoadBalancers, nil
}

// TODO add validation
func (a *APILoadBalancers) Validate() error {
	return nil
}

// FromYaml allows to restore cluster state from YAML.
func FromYaml(c []byte) (*apiLoadBalancers, error) {
	apiLoadBalancers := &APILoadBalancers{}
	if err := yaml.Unmarshal(c, &apiLoadBalancers); err != nil {
		return nil, fmt.Errorf("failed to parse input yaml: %w", err)
	}
	p, err := apiLoadBalancers.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster object: %w", err)
	}
	return p, nil
}

// StateToYaml allows to dump cluster state to YAML, so it can be restored later.
func (a *apiLoadBalancers) StateToYaml() ([]byte, error) {
	return yaml.Marshal(APILoadBalancers{State: a.containers.PreviousState})
}

func (a *apiLoadBalancers) CheckCurrentState() error {
	containers, err := a.containers.New()
	if err != nil {
		return err
	}
	if err := containers.CheckCurrentState(); err != nil {
		return err
	}
	a.containers = *containers.ToExported()
	return nil
}

func (a *apiLoadBalancers) Deploy() error {
	containers, err := a.containers.New()
	if err != nil {
		return err
	}
	// TODO Deploy shouldn't refresh the state. However, due to how we handle exported/unexported
	// structs to enforce validation of objects, we lose current state, as we want it to be computed.
	// On the other hand, maybe it's a good thing to call it once we execute. This way we could compare
	// the plan user agreed to execute with plan calculated right before the execution and fail early if they
	// differ.
	// This is similar to what terraform is doing and may cause planning to run several times, so it may require
	// some optimization.
	// Alternatively we can have serializable plan and a knob in execute command to control whether we should
	// make additional validation or not.
	if err := containers.CheckCurrentState(); err != nil {
		return err
	}
	if err := containers.Execute(); err != nil {
		return err
	}
	a.containers = *containers.ToExported()
	return nil
}
