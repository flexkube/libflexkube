// +build integration

package step

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

// Apply()
func TestAddNodeApply(t *testing.T) {
	node := &node.Node{
		Name:  "foo",
		Image: "gcr.io/etcd-development/etcd:v3.3.13",
	}
	step, err := AddNodeStep(node)
	if err != nil {
		t.Errorf("Step with valid node should be created")
	}
	node, err = step.Apply()
	if err != nil {
		t.Errorf("Step should be applied succesfully")
	}
}

func TestAddNodeApplyBadImage(t *testing.T) {
	node := &node.Node{
		Name:  "foo",
		Image: "gcr.io/etcd-development/etcd:v3.3.13",
	}
	step, err := AddNodeStep(node)
	if err != nil {
		t.Errorf("Step with valid node should be created")
	}
	step.Node.Image = "foo"
	node, err = step.Apply()
	if err == nil {
		t.Errorf("Step with bad image name should fails")
	}
}
