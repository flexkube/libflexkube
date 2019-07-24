package etcd

import "testing"

func TestAddNode(t *testing.T) {
	etcd := New()
	node := "foo"
	if err := etcd.DesiredState.AddNode(node); err != nil {
		t.Errorf("Failed to add '%s' node ", node)
	}
	if etcd.DesiredState.Nodes[node] == nil {
		t.Errorf("Node %s has not been added", node)
	}
}

func TestDuplicatedNode(t *testing.T) {
	etcd := New()
	node := "foo"
	if err := etcd.DesiredState.AddNode(node); err != nil {
		t.Errorf("Failed to add '%s' node ", node)
	}
	if err := etcd.DesiredState.AddNode(node); err == nil {
		t.Errorf("Duplicated node '%s' has been added", node)
	}
}
