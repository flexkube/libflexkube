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

// APILoadBalancers allows to manage group of kube-apiserver load balancer containers, which
// can be used to build highly available Kubernetes cluster.
//
// The main use case is to create an load balancer instance in front of the kubelet on every
// node, as kubelet itself does not support failover for configured API server address.
//
// Current implementation uses HAProxy for load balancing with active health checking, so if
// one of API servers go down, it won't be used by the kubelet. Otherwise some of kubelet requests
// could timeout, hitting unreachable API server.
//
// API load balancer can be also used to expose Kubernetes API to the internet, if it is only
// available in private network.
//
// The HAProxy is configured to run in TCP mode, so potential performance and security overhead
// should be negligible.
type APILoadBalancers struct {
	// Image allows to set Docker image with tag, which will be used by all instances,
	// if instance itself has no image set. If empty, haproxy image defined in pkg/defaults
	// will be used.
	//
	// Example value: 'haproxy:2.1.4-alpine'
	//
	// This field is optional.
	Image string `json:"image,omitempty"`

	// SSH stores common SSH configuration for all instances and will be merged with instances
	// SSH configuration. If instance has some SSH fields defined, they take precedence over
	// this block.
	//
	// If you use same username or port for all members, it is recommended to have it defined
	// here to avoid repetition in the configuration.
	//
	// This field is optional.
	SSH *ssh.Config `json:"ssh,omitempty"`

	// Servers is a list of Kubernetes API server addresses, which should be used as a backend
	// servers.
	//
	// Example value: '[]string{"192.168.10.10:6443", "192.168.10.11:6443"}'.
	//
	// If specified, this value will be used for all instances, which do not have it defined.
	//
	// This field is optional.
	Servers []string `json:"servers,omitempty"`

	// APILoadBalancers is a list of load balancer instances to create. Usually it has only
	// instance specific information defined, like IP address to listen on or on which host
	// the container should be created. See APILoadBalancer struct to see available fields.
	//
	// If there is no state defined, this list must not be empty.
	//
	// If state is defined and list is empty, all created containers will be removed.
	APILoadBalancers []APILoadBalancer `json:"apiLoadBalancers,omitempty"`

	// Name is a container name to create. If you want to run more than one instance on a
	// single host, this field must be unique, otherwise you will get an error with duplicated
	// container name.
	//
	// Currently, when you want to expose Kubernetes API using the load balancer on more than
	// one specific address (so not listening on 0.0.0.0), then two pools are required.
	// This limitation will be addressed in the future.
	//
	// If specified, this value will be used for all instances, which do not have it defined.
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
	// If specified, this value will be used for all instances, which do not have it defined.
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
	// If specified, this value will be used for all instances, which do not have it defined.
	//
	// This field is optional.
	BindAddress string `json:"bindAddress,omitempty"`

	// State stores state of the created containers. After deployment, it is up to the user to export
	// the state and restore it on consecutive runs.
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
		return nil, fmt.Errorf("validating API Load balancers configuration: %w", err)
	}

	cc := &container.Containers{
		PreviousState: a.State,
		DesiredState:  container.ContainersState{},
	}

	for i, lb := range a.APILoadBalancers {
		lb := lb
		a.propagateInstance(&lb)

		lbx, _ := lb.New()                           //nolint:errcheck // Already checked in Validate().
		lbxHcc, _ := lbx.ToHostConfiguredContainer() //nolint:errcheck // Already checked in Validate().

		cc.DesiredState[strconv.Itoa(i)] = lbxHcc
	}

	c, _ := cc.New() //nolint:errcheck // Already checked in Validate().

	return &apiLoadBalancers{
		containers: c,
	}, nil
}

// Validate validates APILoadBalancers struct.
func (a *APILoadBalancers) Validate() error {
	var errors util.ValidateErrors

	cc := &container.Containers{
		PreviousState: a.State,
		DesiredState:  container.ContainersState{},
	}

	for i, lb := range a.APILoadBalancers {
		lb := lb
		a.propagateInstance(&lb)

		lbx, err := lb.New()
		if err != nil {
			errors = append(errors, fmt.Errorf("creating load balancer instance %q: %w", i, err))

			continue
		}

		lbxHcc, err := lbx.ToHostConfiguredContainer()
		if err != nil {
			errors = append(errors, fmt.Errorf("creating load balancer %q container configuration: %w", i, err))

			continue
		}

		cc.DesiredState[strconv.Itoa(i)] = lbxHcc
	}

	noContainersDefined := len(a.State) == 0 && len(a.APILoadBalancers) == 0
	if noContainersDefined {
		errors = append(errors, fmt.Errorf("at least one load balancer must be defined if state is empty"))
	}

	if _, err := cc.New(); !noContainersDefined && err != nil {
		errors = append(errors, fmt.Errorf("creating containers object: %w", err))
	}

	return errors.Return()
}

// FromYaml allows to create and validate resource from YAML format.
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
