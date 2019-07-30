package step

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

// RemoveNodeStep()
func TestRemoveNodeStepValidateNode(t *testing.T) {
	node := &node.Node{}
	if _, err := RemoveNodeStep(node); err == nil {
		t.Errorf("Invalid node should be rejected")
	}
}

func TestRemoveNodeStep(t *testing.T) {
	node := &node.Node{
		Name:  "foo",
		Image: "gcr.io/etcd-development/etcd:v3.3.13",
	}
	_, err := RemoveNodeStep(node)
	if err != nil {
		t.Errorf("Step with valid node should be created")
	}
}

// String()
func TestDescribeRemoveNode(t *testing.T) {
	step := &RemoveNode{
		Node: &node.Node{
			Name: "foo",
		},
	}
	if _, err := step.String(); err != nil {
		t.Errorf("Describing RemoveNode step should pass, got: %s", err)
	}
}

// Validate()
func TestRemoveNodeStepValidate(t *testing.T) {
	node := &node.Node{
		Name:  "foo",
		Image: "gcr.io/etcd-development/etcd:v3.3.13",
	}
	step, err := RemoveNodeStep(node)
	if err != nil {
		t.Errorf("Step with valid node should be created")
	}
	if err := step.Validate(); err != nil {
		t.Errorf("Step should be valid, got: '%s'", err)
	}
}

func TestRemoveNodeStepValidateNoNode(t *testing.T) {
	node := &node.Node{
		Name:  "foo",
		Image: "gcr.io/etcd-development/etcd:v3.3.13",
	}
	step, err := RemoveNodeStep(node)
	if err != nil {
		t.Errorf("Step with valid node should be created")
	}
	step.Node = nil
	if err := step.Validate(); err == nil {
		t.Errorf("Step without node should not be valid")
	}
}

func TestRemoveNodeStepValidateInvalidNode(t *testing.T) {
	node := &node.Node{
		Name:  "foo",
		Image: "gcr.io/etcd-development/etcd:v3.3.13",
	}
	step, err := RemoveNodeStep(node)
	if err != nil {
		t.Errorf("Step with valid node should be created, got '%s'", err)
	}
	step.Node.Name = ""
	if err := step.Validate(); err == nil {
		t.Errorf("Step with invalid node should not be valid")
	}
}
