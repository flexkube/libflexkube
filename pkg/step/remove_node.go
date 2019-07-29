package step

import (
	"fmt"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

// RemoveNodeStep creates new RemoveNode step and validates it
//
// Takes node which should be removed from the cluster as argument
func RemoveNodeStep(node *node.Node) (*Step, error) {
	if err := node.Validate(); err != nil {
		return nil, fmt.Errorf("node not valid: %s", err)
	}
	return &Step{
		Type: RemoveNode,
		Node: node,
	}, nil
}
