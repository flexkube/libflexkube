package etcd

import "testing"

func TestAddNode(t *testing.T) {
	state := NewState()
	node := "foo"
	if err := state.AddNode(node); err != nil {
		t.Errorf("Failed to add '%s' node ", node)
	}
	if state.Nodes[node] == nil {
		t.Errorf("Node %s has not been added", node)
	}
}

func TestDuplicatedNode(t *testing.T) {
	state := NewState()
	node := "foo"
	if err := state.AddNode(node); err != nil {
		t.Errorf("Failed to add '%s' node ", node)
	}
	if err := state.AddNode(node); err == nil {
		t.Errorf("Duplicated node '%s' has been added", node)
	}
}

func TestRemoveNode(t *testing.T) {
	state := NewState()
	node := "foo"
	err := state.AddNode(node)
	if err = state.RemoveNode(node); err != nil {
		t.Errorf("Node should be removed")
	}
}

func TestRemoveNonExistingNode(t *testing.T) {
	state := NewState()
	node := "foo"
	if err := state.RemoveNode(node); err == nil {
		t.Errorf("Node which do not exist should not be removed")
	}
}
