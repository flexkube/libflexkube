package step

import (
	"fmt"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

// AddNodeStep creates new AddNode step and validates it
//
// Takes node struct which should be added as an argument
func AddNodeStep(node *node.Node) (*Step, error) {
	if err := node.Validate(); err != nil {
		return nil, fmt.Errorf("node not valid: %s", err)
	}
	return &Step{
		StepType: AddNode,
		Node:     node,
	}, nil
}

// DescribeAddNode returns human readable description of AddNode step
func (step *Step) DescribeAddNode() (string, error) {
	if step.StepType != AddNode {
		return "", fmt.Errorf("wrong step type, expected AddNode, got '%s'", step.StepType)
	}
	if step.Node.Image == "" {
		return fmt.Sprintf("Add node '%s' with unknown image", step.Node.Name), nil
	}
	return fmt.Sprintf("Add node '%s' with image '%s'", step.Node.Name, step.Node.Image), nil
}

func (step *Step) ValidateAddNode() error {
	if step.StepType != AddNode {
		return fmt.Errorf("wrong step type, expected AddNode, got '%s'", step.StepType)
	}

	if step.Node == nil {
		return fmt.Errorf("AddNode step should have node field set")
	}
	if err := step.Node.Validate(); err != nil {
		return fmt.Errorf("AddNode step contains invalid node object")
	}

	return nil
}
