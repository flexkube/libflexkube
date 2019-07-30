// +build integration

package etcd

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

func TestExecutePlanDocker(t *testing.T) {
	etcd := New()
	node := &node.Node{
		Name:  "foo",
		Image: "gcr.io/etcd-development/etcd:v3.3.13",
	}
	if err := etcd.AddNode(node); err != nil {
		t.Errorf("Adding new node should not fail, got: %s", err)
	}
	if err := etcd.ReadCurrentState(); err != nil {
		t.Errorf("Reading state of empty cluster should succeed, got: %s", err)
	}
	if err := etcd.Plan(); err != nil {
		t.Errorf("Planning should succeed, got: %s", err)
	}
	if err := etcd.Apply(); err != nil {
		t.Errorf("Applying should succeed, got: %s", err)
	}
}
