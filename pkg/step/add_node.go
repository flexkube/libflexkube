package step

import (
	"fmt"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container"
	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
	"github.com/pkg/errors"
)

// AddNode represent step, which adds new node to the cluster
type AddNode struct {
	Node *node.Node
}

// AddNodeStep creates new AddNode step and validates it
//
// Takes node struct which should be added as an argument
func AddNodeStep(node *node.Node) (*AddNode, error) {
	if err := node.Validate(); err != nil {
		return nil, fmt.Errorf("node not valid: %s", err)
	}
	return &AddNode{
		Node: node,
	}, nil
}

// String returns human readable description of AddNode step
func (step *AddNode) String() (string, error) {
	if step.Node.Image == "" {
		return fmt.Sprintf("Add node '%s' with unknown image", step.Node.Name), nil
	}
	return fmt.Sprintf("Add node '%s' with image '%s'", step.Node.Name, step.Node.Image), nil
}

// Validate validates step of type AddNode
func (step *AddNode) Validate() error {
	if step.Node == nil {
		return fmt.Errorf("AddNode step should have node field set")
	}
	if err := step.Node.Validate(); err != nil {
		return fmt.Errorf("AddNode step contains invalid node object")
	}

	return nil
}

// Apply adds node to the cluster, reads it state and returns it
func (step *AddNode) Apply() (*node.Node, error) {
	c, err := step.Node.SelectContainerRuntime()
	if err != nil {
		return nil, fmt.Errorf("Unable to determine container runtime: %s", err)
	}
	config := &container.Config{
		Name:  step.Node.Name,
		Image: step.Node.Image,
	}
	if err := c.Start(config); err != nil {
		return nil, errors.Wrap(err, "starting container")
	}
	return step.Node, nil
}
