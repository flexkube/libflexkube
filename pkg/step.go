package etcd

import (
	"fmt"
)

type StepType int

const (
	AddNode = iota
)

// Step describes single action of etcd cluster modification
type Step struct {
	Description string
	StepType    StepType
	Node        *Node
}

func AddNodeStep(node *Node) (*Step, error) {
	if err := node.Validate(); err != nil {
		return nil, fmt.Errorf("node not valid: %s", err)
	}
	return &Step{
		StepType: AddNode,
		Node:     node,
	}, nil
}

func (step *Step) Describe() (string, error) {
	if err := step.Validate(); err != nil {
		return "", fmt.Errorf("Unable to describe invalid step")
	}
	switch step.StepType {
	case AddNode:
		if step.Node.Image == "" {
			return fmt.Sprintf("Add node '%s' with unknown image", step.Node.Name), nil
		}
		return fmt.Sprintf("Add node '%s' with image '%s'", step.Node.Name, step.Node.Image), nil
	default:
		return "", fmt.Errorf("Unknown step")
	}
}

func (step *Step) Validate() error {
	switch step.StepType {
	case AddNode:
		if step.Node == nil {
			return fmt.Errorf("AddNode step should have node field set")
		}
		if err := step.Node.Validate(); err != nil {
			return fmt.Errorf("AddNode step contains invalid node object")
		}
	default:
		return fmt.Errorf("Unable to validate unknown step")
	}
	return nil
}
