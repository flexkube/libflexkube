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
	Image            string            `json:"image" yaml:"image"`
	SSH              *ssh.Config       `json:"ssh" yaml:"ssh"`
	Servers          []string          `json:"servers" yaml:"servers"`
	APILoadBalancers []APILoadBalancer `json:"apiLoadBalancers" yaml:"apiLoadBalancers"`
	BindPort         int               `json:"bindPort" yaml:"bindPort"`

	// Serializable fields
	State container.ContainersState `json:"state:omitempty" yaml:"state"`
}

// apiLoadBalancers is validated and executable version of APILoadBalancers
type apiLoadBalancers struct {
	image      string
	ssh        *ssh.Config
	containers *container.Containers
}

func (a *APILoadBalancers) propagateInstance(i *APILoadBalancer) {
	i.Image = util.PickString(i.Image, a.Image)
	i.Servers = util.PickStringSlice(i.Servers, a.Servers)
	i.BindPort = util.PickInt(i.BindPort, a.BindPort)
	i.Host = host.BuildConfig(i.Host, host.Host{
		SSHConfig: a.SSH,
	})
}

// New validates APILoadBalancers struct and fills all required fields in members with default values
// provided by the user.
//
// TODO move filling the defaults to separated function, so it can be re-used in Validate.
func (a *APILoadBalancers) New() (types.Resource, error) {
	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate API Load balancers configuration: %w", err)
	}

	apiLoadBalancers := &apiLoadBalancers{
		image: a.Image,
		ssh:   a.SSH,
		containers: &container.Containers{
			PreviousState: a.State,
			DesiredState:  make(container.ContainersState),
		},
	}

	for i, lb := range a.APILoadBalancers {
		lb := lb
		a.propagateInstance(&lb)

		lbx, _ := lb.New()

		apiLoadBalancers.containers.DesiredState[strconv.Itoa(i)] = lbx.ToHostConfiguredContainer()
	}

	return apiLoadBalancers, nil
}

// Validate validates APILoadBalancers struct.
func (a *APILoadBalancers) Validate() error {
	for _, lb := range a.APILoadBalancers {
		lb := lb
		a.propagateInstance(&lb)

		if _, err := lb.New(); err != nil {
			return fmt.Errorf("failed creating load balancer instance: %w", err)
		}
	}

	return nil
}

// FromYaml allows to restore cluster state from YAML.
func FromYaml(c []byte) (types.Resource, error) {
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

// CheckCurrentState reads current state of the deployed resources.
func (a *apiLoadBalancers) CheckCurrentState() error {
	return a.containers.CheckCurrentState()
}

// Deploy checks current status of deployed group of instances and updates them if there is some
// configuration drift.
func (a *apiLoadBalancers) Deploy() error {
	return a.containers.Deploy()
}
