// +build integration

package step

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

// Apply()
func TestRemoveNodeApply(t *testing.T) {
	node := &node.Node{
		Name: "foo",
	}
	step, err := RemoveNodeStep(node)
	if err != nil {
		t.Errorf("Step with valid node should be created")
	}
	node, err = step.Apply()
	if err != nil {
		t.Errorf("Step should be applied succesfully")
	}
}
