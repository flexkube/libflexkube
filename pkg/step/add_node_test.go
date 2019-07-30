package step

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

// AddNodeStep()
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

// String()
func TestDescribeAddNode(t *testing.T) {
	step := &AddNode{
		Node: &node.Node{
			Name: "foo",
		},
	}
	if _, err := step.String(); err != nil {
		t.Errorf("Describing AddNode step should pass, got: %s", err)
	}
}

func TestDescribeAddNodeWithImage(t *testing.T) {
	step := &AddNode{
		Node: &node.Node{
			Name:  "foo",
			Image: "bar",
		},
	}
	if _, err := step.String(); err != nil {
		t.Errorf("Describing AddNode step should pass, got: %s", err)
	}
}

// Validate()
func TestAddNodeStepValidate(t *testing.T) {
	node := &node.Node{
		Name:  "foo",
		Image: "gcr.io/etcd-development/etcd:v3.3.13",
	}
	step, err := AddNodeStep(node)
	if err != nil {
		t.Errorf("Step with valid node should be created")
	}
	if err := step.Validate(); err != nil {
		t.Errorf("Step should be valid, got: '%s'", err)
	}
}

func TestAddNodeStepValidateNoNode(t *testing.T) {
	node := &node.Node{
		Name:  "foo",
		Image: "gcr.io/etcd-development/etcd:v3.3.13",
	}
	step, err := AddNodeStep(node)
	if err != nil {
		t.Errorf("Step with valid node should be created")
	}
	step.Node = nil
	if err := step.Validate(); err == nil {
		t.Errorf("Step without node should not be valid")
	}
}

func TestAddNodeStepValidateInvalidNode(t *testing.T) {
	node := &node.Node{
		Name:  "foo",
		Image: "gcr.io/etcd-development/etcd:v3.3.13",
	}
	step, err := AddNodeStep(node)
	if err != nil {
		t.Errorf("Step with valid node should be created, got '%s'", err)
	}
	step.Node.Name = ""
	if err := step.Validate(); err == nil {
		t.Errorf("Step with invalid node should not be valid")
	}
}
