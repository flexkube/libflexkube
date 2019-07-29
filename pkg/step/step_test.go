package step

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

func TestAddNodeStepValidateInvalid(t *testing.T) {
	step := &Step{
		Type: AddNode,
	}
	if err := step.Validate(); err == nil {
		t.Errorf("Invalid step should't be valid")
	}
}

func TestAddNodeStepValidateInvalidNode(t *testing.T) {
	step := &Step{
		Type: AddNode,
		Node: &node.Node{},
	}
	if err := step.Validate(); err == nil {
		t.Errorf("Invalid step should't be valid")
	}
}

func TestAddNodeStepValidateValid(t *testing.T) {
	step := &Step{
		Type: AddNode,
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
		Type: 99,
	}
	if err := step.Validate(); err == nil {
		t.Errorf("Validating unknown step should fail")
	}
}

func TestDescribeUnknownStep(t *testing.T) {
	step := &Step{
		Type: 99,
	}
	if _, err := step.Describe(); err == nil {
		t.Errorf("Describing unknown step should fail")
	}
}

func TestDescribeValidAddNodeStep(t *testing.T) {
	step := &Step{
		Type: AddNode,
		Node: &node.Node{
			Name: "foo",
		},
	}
	if _, err := step.Describe(); err != nil {
		t.Errorf("Valid step should be described, got: %s", err)
	}
}
