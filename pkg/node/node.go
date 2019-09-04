package node

import (
	"fmt"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container"
	"github.com/invidian/etcd-ariadnes-thread/pkg/container/docker"

	"github.com/pkg/errors"
)

// Node represents public, serializable version of the node object.
//
// It should be used for persisting and restoring node state with combination
// with New(), which make sure that the configuration is actually correct.
type Node struct {
	// Name is an unique indentier for the node.
	Name string `json:"name"`
	// Image is a Docker image name with tag of the node.
	Image string `json:"image"`
	// ContainerName is a name which will be used for the container
	ContainerName string `json:"container_name"`
	// Status of the container
	ContainerStatus *container.Status `json:"container_status"`
	// Name of the container runtime to use. If not specified, container runtime
	// will be automatically determined based on configuration options. If no configuration
	// options are specified, "docker" will be chosen.
	ContainerRuntimeName string `json:"container_runtime_name,omit_empty"`
}

// node represents validated version of Node object, which contains all requires
// information for instantiating (by calling Create()).
type node struct {
	// Contains common information between node and nodeInstance
	nodeBase
	// Optional container status
	containerStatus *container.Status
}

// node represents created container. It guarantees that container status is initialised.
type nodeInstance struct {
	// Contains common information between node and nodeInstance
	nodeBase
	// Status of the container
	containerStatus container.Status
}

// nodeBase contains basic information about created and not created node.
type nodeBase struct {
	// name is an unique indentier for the node.
	name string
	// Image is a Docker image name with tag of the node.
	image string
	// ContainerName is a node name combined with cluster name
	containerName string
	// Container runtime which will be used to manage the container
	containerRuntime container.Container
	// name of container runtime to use
	containerRuntimeName string
}

// New creates new instance of node from Node and validates it's configuration
func New(n *Node) (*node, error) {
	if err := n.Validate(); err != nil {
		return nil, errors.Wrap(err, "node configuration validation failed")
	}
	nn := &node{
		nodeBase{
			name:                 n.Name,
			image:                n.Image,
			containerName:        n.ContainerName,
			containerRuntimeName: n.ContainerRuntimeName,
		},
		n.ContainerStatus,
	}
	if err := nn.selectContainerRuntime(); err != nil {
		return nil, errors.Wrap(err, "unable to determine container runtime")
	}
	return nn, nil
}

// Validate validates node configuration
func (n *Node) Validate() error {
	if n.Name == "" {
		return fmt.Errorf("name must be set")
	}

	if n.ContainerName == "" {
		return fmt.Errorf("container name must be set")
	}

	if !container.IsRuntimeSupported(n.ContainerRuntimeName) {
		return fmt.Errorf("configured runtime '%s' is not supported", n.ContainerRuntimeName)
	}

	return nil
}

// selectContainerRuntime returns container runtime configured for node
//
// It returns error if container runtime configuration is invalid
func (node *node) selectContainerRuntime() error {
	switch container.GetRuntimeName(node.containerRuntimeName) {
	case "docker":
		c, err := docker.New(&docker.Docker{})
		if err != nil {
			return errors.Wrap(err, "selecting container runtime failed")
		}
		node.containerRuntime = c
	default:
		return fmt.Errorf("not supported container runtime: %s", node.containerRuntimeName)
	}
	return nil
}

// Create creates node container from it's defintion
func (n *node) Create() (*nodeInstance, error) {
	config := &container.Config{
		Name:  n.containerName,
		Image: n.image,
	}
	id, err := n.containerRuntime.Create(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating container")
	}

	return &nodeInstance{
		nodeBase{
			name:                 n.name,
			image:                n.image,
			containerName:        n.containerName,
			containerRuntime:     n.containerRuntime,
			containerRuntimeName: n.containerRuntimeName,
		},
		container.Status{
			ID: id,
		},
	}, nil
}

// FromStatus creates nodeInstance from previously restored status
func (n *node) FromStatus() (*nodeInstance, error) {
	if n.containerStatus == nil || n.containerStatus.ID == "" {
		return nil, fmt.Errorf("can't create node instance from invalid status")
	}
	return &nodeInstance{
		nodeBase:        n.nodeBase,
		containerStatus: *n.containerStatus,
	}, nil
}

// ReadState reads state of the node from container runtime
//
// TODO Should we store the status here or return it instead?
func (node *nodeInstance) Status() (*container.Status, error) {
	status, err := node.containerRuntime.Status(node.containerStatus.ID)
	if err != nil {
		return nil, errors.Wrap(err, "getting container status failed")
	}
	if status == nil {
		return nil, fmt.Errorf("container '%s' does not exist", node.containerStatus.ID)
	}

	node.containerStatus = *status

	return status, nil
}

// Start starts the node container
func (node *nodeInstance) Start() error {
	return node.containerRuntime.Start(node.containerStatus.ID)
}

// Stop stops the node container
func (node *nodeInstance) Stop() error {
	return node.containerRuntime.Stop(node.containerStatus.ID)
}

// Delete removes node container
//
// TODO should we also clean up host directories here?
func (node *nodeInstance) Delete() error {
	return node.containerRuntime.Delete(node.containerStatus.ID)
}
