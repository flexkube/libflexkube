package step

import (
	"fmt"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

// RemoveNode represents step, which removes existing node from the cluster
type RemoveNode struct {
	Node *node.Node
}

// RemoveNodeStep creates new RemoveNode step and validates it
//
// Takes node which should be removed from the cluster as argument
func RemoveNodeStep(node *node.Node) (*RemoveNode, error) {
	if err := node.Validate(); err != nil {
		return nil, fmt.Errorf("node not valid: %s", err)
	}
	return &RemoveNode{
		Node: node,
	}, nil
}

// String returns human readable description of RemoveNode step
func (step *RemoveNode) String() (string, error) {
	return fmt.Sprintf("Remove node '%s' from the cluster", step.Node.Name), nil
}

// Validate validates step of type RemoveNode
func (step *RemoveNode) Validate() error {
	if step.Node == nil {
		return fmt.Errorf("RemoveNode step should have node field set")
	}
	if err := step.Node.Validate(); err != nil {
		return fmt.Errorf("RemoveNode step contains invalid node object")
	}

	return nil
}

// Apply removes node from the cluster
func (step *RemoveNode) Apply() (*node.Node, error) {
	return step.Node, nil
}
