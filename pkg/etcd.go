package etcd

import "fmt"

type Etcd struct {
	Nodes map[string]*EtcdNode
}

type EtcdNode struct {
	Name string
}

func New() *Etcd {
	return &Etcd{
		Nodes: make(map[string]*EtcdNode),
	}
}

func (etcd *Etcd) AddNode(name string) error {
	if etcd.Nodes[name] != nil {
		return fmt.Errorf("Node already exists")
	}
	etcd.Nodes[name] = &EtcdNode{
		Name: name,
	}

	return nil
}
