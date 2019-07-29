package step

import (
	"fmt"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

// Type is Go method for clear enums
type Type int

// List of possible steps
// TODO consider adding backup/restore steps?
const (
	// AddNode step adds new node to the cluster
	AddNode = iota
	// RemoveNode gracefully removes node from the cluster
	RemoveNode
	// UpdateNode updates single node configuration/image
	UpdateNode
	// HealNode attempts to restore node to a healthy state
	HealNode
)

func (d Type) String() string {
	return [...]string{"AddNode", "RemoveNode", "UpdateNode", "HealNode"}[d]
}

// Steps is an alias for storing multiple steps
type Steps []*Step

// Step describes single action of etcd cluster modification
type Step struct {
	Description string
	Type        Type
	Node        *node.Node
}

// Describe creates human readable description about the step
func (step *Step) Describe() (string, error) {
	if err := step.Validate(); err != nil {
		return "", fmt.Errorf("Unable to describe invalid step")
	}
	switch step.Type {
	case AddNode:
		return step.DescribeAddNode()
	default:
		return "", fmt.Errorf("Unknown step")
	}
}

// Validate makes sure the step is valid
func (step *Step) Validate() error {
	switch step.Type {
	case AddNode:
		return step.ValidateAddNode()
	default:
		return fmt.Errorf("Unable to validate unknown step")
	}
}
