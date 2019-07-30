package node

import (
	"fmt"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container"
	"github.com/invidian/etcd-ariadnes-thread/pkg/container/docker"
)

// Node represents cluster member, information how it should be set up, how
// to connect to it etc.
type Node struct {
	// Unique indentier for the node.
	Name string
	// Docker image name with tag of the node.
	Image string
}

// ReadState reads general state of the node.
func (node *Node) ReadState() error {
	if err := node.ReadImage(); err != nil {
		return err
	}
	return nil
}

// ReadImage reads which image is currently running on existing deployment
// and sets it in the node object.
func (node *Node) ReadImage() error {
	node.Image = "notimplemented"
	return nil
}

// Validate valdiates node configuration
func (node *Node) Validate() error {
	if node.Name == "" {
		return fmt.Errorf("name must be set")
	}

	return nil
}

// SelectContainerRuntime returns container runtime configured for node
func (node *Node) SelectContainerRuntime() (container.Container, error) {
	return &docker.Docker{}, nil
}
