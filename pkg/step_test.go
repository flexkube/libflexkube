package etcd

import "testing"

func TestAddNodeStepValidateNode(t *testing.T) {
	node := &Node{}
	if _, err := AddNodeStep(node); err == nil {
		t.Errorf("Invalid node should be rejected")
	}
}

func TestAddNodeStep(t *testing.T) {
	node := &Node{
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
		Node:     &Node{},
	}
	if err := step.Validate(); err == nil {
		t.Errorf("Invalid step should't be valid")
	}
}

func TestAddNodeStepValidateValid(t *testing.T) {
	step := &Step{
		StepType: AddNode,
		Node: &Node{
			Name: "foo",
		},
	}
	if err := step.Validate(); err != nil {
		t.Errorf("Valid step should be valid")
	}
}
