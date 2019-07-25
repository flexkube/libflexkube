package etcd

import "fmt"

type EtcdState struct {
	Nodes map[string]*EtcdNode
}

type EtcdNode struct {
	Name string
}

func NewState() *EtcdState {
	nodes := make(map[string]*EtcdNode)
	return &EtcdState{
		Nodes: nodes,
	}
}

func (state *EtcdState) AddNode(name string) error {
	if state.Nodes[name] != nil {
		return fmt.Errorf("Node already exists")
	}
	state.Nodes[name] = &EtcdNode{
		Name: name,
	}

	return nil
}
