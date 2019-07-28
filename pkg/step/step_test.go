package step

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

func TestAddNodeStepValidateNode(t *testing.T) {
	node := &node.Node{}
	if _, err := AddNodeStep(node); err == nil {
		t.Errorf("Invalid node should be rejected")
	}
}

func TestAddNodeStep(t *testing.T) {
	node := &node.Node{
		Name:  "foo",
		Image: "gcr.io/etcd-development/etcd:v3.3.13",
	}
	if _, err := AddNodeStep(node); err != nil {
		t.Errorf("Step with valid node should be created")
	}
}

func TestAddNodeStepValidateInvalid(t *testing.T) {
	step := &Step{
		StepType: AddNode,
	}
	if err := step.Validate(); err == nil {
		t.Errorf("Invalid step should't be valid")
	}
}

func TestAddNodeStepValidateInvalidNode(t *testing.T) {
	step := &Step{
		StepType: AddNode,
		Node:     &node.Node{},
	}
	if err := step.Validate(); err == nil {
		t.Errorf("Invalid step should't be valid")
	}
}

func TestAddNodeStepValidateValid(t *testing.T) {
	step := &Step{
		StepType: AddNode,
		Node: &node.Node{
			Name: "foo",
		},
	}
	if err := step.Validate(); err != nil {
		t.Errorf("Valid step should be valid")
	}
}

func TestValidateUnknownStep(t *testing.T) {
	step := &Step{
		StepType: 99,
	}
	if err := step.Validate(); err == nil {
		t.Errorf("Validating unknown step should fail")
	}
}

func TestDescribeUnknownStep(t *testing.T) {
	step := &Step{
		StepType: 99,
	}
	if _, err := step.Describe(); err == nil {
		t.Errorf("Describing unknown step should fail")
	}
}

func TestDescribeValidAddNodeStep(t *testing.T) {
	step := &Step{
		StepType: AddNode,
		Node: &node.Node{
			Name: "foo",
		},
	}
	if _, err := step.Describe(); err != nil {
		t.Errorf("Valid step should be described, got: %s", err)
	}
}
