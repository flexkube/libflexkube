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

func TestDescribeAddNodeInvalidType(t *testing.T) {
	step := &Step{
		Type: RemoveNode,
	}
	if _, err := step.DescribeAddNode(); err == nil {
		t.Errorf("Invalid step type shouldn't be described")
	}
}

func TestValidateAddNodeInvalidType(t *testing.T) {
	step := &Step{
		Type: RemoveNode,
	}
	if err := step.ValidateAddNode(); err == nil {
		t.Errorf("Invalid step type shouldn't be valid")
	}
}
