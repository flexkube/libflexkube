package etcd

import (
	"testing"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

func TestAddNode(t *testing.T) {
	state := NewState()
	node := &node.Node{
		Name: "foo",
	}
	if err := state.AddNode(node); err != nil {
		t.Errorf("Failed to add '%s' node ", node)
	}
	if state.Nodes[node.Name] == nil {
		t.Errorf("Node %s has not been added", node.Name)
	}
}

func TestDuplicatedNode(t *testing.T) {
	state := NewState()
	node := &node.Node{
		Name: "foo",
	}
	if err := state.AddNode(node); err != nil {
		t.Errorf("Failed to add '%s' node ", node.Name)
	}
	if err := state.AddNode(node); err == nil {
		t.Errorf("Duplicated node '%s' has been added", node.Name)
	}
}

func TestRemoveNode(t *testing.T) {
	state := NewState()
	node := &node.Node{
		Name: "foo",
	}
	if err := state.AddNode(node); err != nil {
		t.Errorf("Adding ndoe should not fail, got: %s", err)
	}
	if err := state.RemoveNode(node.Name); err != nil {
		t.Errorf("Node should be removed, got: %s", err)
	}
}

func TestRemoveNonExistingNode(t *testing.T) {
	state := NewState()
	node := "foo"
	if err := state.RemoveNode(node); err == nil {
		t.Errorf("Node which do not exist should not be removed")
	}
}

func TestAddNodeDefaultImage(t *testing.T) {
	state := NewState()
	imageName := "gcr.io/etcd-development/etcd:v3.3.13"
	state.Image = imageName
	nodeName := "foo"
	node := &node.Node{
		Name: nodeName,
	}
	if err := state.AddNode(node); err != nil {
		t.Errorf("Adding node should work")
	}
	if state.Nodes[nodeName].Image != imageName {
		t.Errorf("node image should be the same as state image '%s', got '%s'", imageName, state.Nodes[nodeName].Image)
	}
}

func TestAddNodeNoDefaultImage(t *testing.T) {
	state := NewState()
	nodeName := "foo"
	node := &node.Node{
		Name: nodeName,
	}
	if err := state.AddNode(node); err != nil {
		t.Errorf("Adding node should work")
	}
	if state.Nodes[nodeName].Image != "" {
		t.Errorf("node image should be empty, got '%s'", state.Nodes[nodeName].Image)
	}
}
