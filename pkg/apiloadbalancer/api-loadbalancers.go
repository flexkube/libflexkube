// Package apiloadbalancer allows to create and manage kube-apiserver load balancer
// containers.
package apiloadbalancer

import (
	"fmt"
	"strconv"

	"sigs.k8s.io/yaml"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container"
	"github.com/flexkube/libflexkube/pkg/host"
	"github.com/flexkube/libflexkube/pkg/host/transport/ssh"
	"github.com/flexkube/libflexkube/pkg/types"
)

// APILoadBalancers represents group of APILoadBalancer instances. It allows to set default values for
// all configured instances.
type APILoadBalancers struct {
	Image            string            `json:"image,omitempty"`
	SSH              *ssh.Config       `json:"ssh,omitempty"`
	Servers          []string          `json:"servers,omitempty"`
	APILoadBalancers []APILoadBalancer `json:"apiLoadBalancers,omitempty"`
	Name             string            `json:"name,omitempty"`
	HostConfigPath   string            `json:"hostConfigPath,omitempty"`
	BindAddress      string            `json:"bindAddress,omitempty"`

	// Serializable fields
	State container.ContainersState `json:"state,omitempty"`
}

// apiLoadBalancers is validated and executable version of APILoadBalancers.
type apiLoadBalancers struct {
	containers container.ContainersInterface
}

func (a *APILoadBalancers) propagateInstance(i *APILoadBalancer) {
	i.Image = util.PickString(i.Image, a.Image)
	i.Servers = util.PickStringSlice(i.Servers, a.Servers)
	i.Host = host.BuildConfig(i.Host, host.Host{
		SSHConfig: a.SSH,
	})
	i.Name = util.PickString(i.Name, a.Name)
	i.HostConfigPath = util.PickString(i.HostConfigPath, a.HostConfigPath)
	i.BindAddress = util.PickString(i.BindAddress, a.BindAddress)
}

// New validates APILoadBalancers struct and fills all required fields in members with default values
// provided by the user.
//
// TODO move filling the defaults to separated function, so it can be re-used in Validate.
func (a *APILoadBalancers) New() (types.Resource, error) {
	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate API Load balancers configuration: %w", err)
	}

	cc := &container.Containers{
		PreviousState: a.State,
		DesiredState:  make(container.ContainersState),
	}

	for i, lb := range a.APILoadBalancers {
		lb := lb
		a.propagateInstance(&lb)

		lbx, _ := lb.New()
		lbxHcc, _ := lbx.ToHostConfiguredContainer()

		cc.DesiredState[strconv.Itoa(i)] = lbxHcc
	}

	c, _ := cc.New()

	return &apiLoadBalancers{
		containers: c,
	}, nil
}

// Validate validates APILoadBalancers struct.
func (a *APILoadBalancers) Validate() error {
	var errors util.ValidateError

	cc := &container.Containers{
		PreviousState: a.State,
		DesiredState:  make(container.ContainersState),
	}

	for i, lb := range a.APILoadBalancers {
		lb := lb
		a.propagateInstance(&lb)

		lbx, err := lb.New()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed creating load balancer instance %q: %w", i, err))
			continue
		}

		lbxHcc, err := lbx.ToHostConfiguredContainer()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed creating load balancer %q container configuration: %w", i, err))
			continue
		}

		cc.DesiredState[strconv.Itoa(i)] = lbxHcc
	}

	noContainersDefined := len(a.State) == 0 && len(a.APILoadBalancers) == 0
	if noContainersDefined {
		errors = append(errors, fmt.Errorf("at least one load balancer must be defined if state is empty"))
	}

	if _, err := cc.New(); !noContainersDefined && err != nil {
		errors = append(errors, fmt.Errorf("failed creating containers object: %w", err))
	}

	return errors.Return()
}

// FromYaml allows to restore cluster state from YAML.
func FromYaml(c []byte) (types.Resource, error) {
	return types.ResourceFromYaml(c, &APILoadBalancers{})
}

// StateToYaml allows to dump cluster state to YAML, so it can be restored later.
func (a *apiLoadBalancers) StateToYaml() ([]byte, error) {
	return yaml.Marshal(APILoadBalancers{State: a.containers.ToExported().PreviousState})
}

// CheckCurrentState reads current state of the deployed resources.
func (a *apiLoadBalancers) CheckCurrentState() error {
	return a.containers.CheckCurrentState()
}

// Deploy checks current status of deployed group of instances and updates them if there is some
// configuration drift.
func (a *apiLoadBalancers) Deploy() error {
	return a.containers.Deploy()
}

// Containers implement types.Resource interface.
func (a *apiLoadBalancers) Containers() container.ContainersInterface {
	return a.containers
}
