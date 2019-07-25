package etcd

import "fmt"

// State contains abstact state of the etcd cluster, which might represent
// both existing, desired and historical state.
type State struct {
	// List of cluster nodes
	Nodes map[string]*Node
	// Image, which should be used by all nodes.
	// If this field is nil, it means cluster consist of 2 or more different images.
	Image string
}

func NewState() *State {
	nodes := make(map[string]*Node)
	return &State{
		Nodes: nodes,
	}
}

func (state *State) AddNode(name string) error {
	if state.Nodes[name] != nil {
		return fmt.Errorf("node already exists")
	}
	state.Nodes[name] = &Node{
		Name: name,
	}

	return nil
}

func (state *State) RemoveNode(name string) error {
	if state.Nodes[name] == nil {
		return fmt.Errorf("node does not exist")
	}
	delete(state.Nodes, name)

	return nil
}

func (state *State) Read() error {
	for _, node := range state.Nodes {
		if err := node.ReadState(); err != nil {
			return err
		}
	}
	return nil
}
