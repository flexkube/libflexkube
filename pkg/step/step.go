package step

import (
	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
)

// Step describes single atomic change which can be applied to the cluster
type Step interface {
	// String produces human-readable description of the step
	String() (string, error)
	// Validate validates step
	Validate() error
	// Apply applies step to the cluster
	Apply() (*node.Node, error)
}

// Steps is an alias for storing multiple steps
type Steps []Step
