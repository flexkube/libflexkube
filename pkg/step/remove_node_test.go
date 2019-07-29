package step

import (
	"reflect"
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

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
	step, err := RemoveNodeStep(node)
	if err != nil {
		t.Errorf("Step with valid node should be created")
	}
	expectedStep := &Step{
		Type: RemoveNode,
		Node: node,
	}
	if !reflect.DeepEqual(expectedStep, step) {
		t.Errorf("Step does not match expected step")
	}
}
