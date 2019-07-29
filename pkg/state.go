package etcd

import (
	"fmt"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

// State contains abstact state of the etcd cluster, which might represent
// both existing, desired and historical state.
//
// TODO State struct should be serializable, so it's easier to persist
type State struct {
	// List of cluster nodes
	Nodes map[string]*node.Node
	// Image, which should be used by all nodes.
	// If this field is nil, it means cluster consist of 2 or more different images.
	Image string
}

// NewState creates new state object with initialized nodes map
func NewState() *State {
	nodes := make(map[string]*node.Node)
	return &State{
		Nodes: nodes,
	}
}

// AddNode validates and adds node object to the state
func (state *State) AddNode(node *node.Node) error {
	if err := node.Validate(); err != nil {
		return fmt.Errorf("node validation failed: %s", err)
	}
	if state.Nodes[node.Name] != nil {
		return fmt.Errorf("node '%s' is already added", node.Name)
	}
	if node.Image == "" && state.Image != "" {
		node.Image = state.Image
	}
	state.Nodes[node.Name] = node

	return nil
}

// RemoveNode removes node from the state. Returns error if node does not exist.
func (state *State) RemoveNode(name string) error {
	if state.Nodes[name] == nil {
		return fmt.Errorf("node does not exist")
	}
	delete(state.Nodes, name)

	return nil
}

// Read reads state of each node
func (state *State) Read() error {
	for _, node := range state.Nodes {
		if err := node.ReadState(); err != nil {
			return err
		}
	}
	return nil
}
