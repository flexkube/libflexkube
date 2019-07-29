package etcd

import (
	"fmt"

	"github.com/invidian/etcd-ariadnes-thread/pkg/node"
	"github.com/invidian/etcd-ariadnes-thread/pkg/step"
)

// Etcd stores all etcd cluster states
type Etcd struct {
	// PreviousState can be loaded from any data store
	// It is required to have previous state to be able to remove old elements
	// This struct is public and can be interacted with from outside
	PreviousState *State
	// CurrentState is loaded based on the PreviousState by doing health checks
	// on each node etc.
	// This struct is private and can be modified only by Etcd functions
	currentState *State
	// DesiredState represents new state requested by user. Deployment process
	// checks what is missing in CurrentState and attempts to fullfil it
	// This struct is public and can be interacted with from outside
	DesiredState *State
	// Steps contains all steps, which should be applied to the cluster.
	Steps step.Steps
}

// New creates new etcd cluster object
func New() *Etcd {
	previousState := NewState()
	desiredState := NewState()

	return &Etcd{
		PreviousState: previousState,
		DesiredState:  desiredState,
	}
}

// AddNode should be used for adding new nodes to the cluster
func (etcd *Etcd) AddNode(node *node.Node) error {
	return etcd.DesiredState.AddNode(node)
}

// ReadCurrentState copies previous state to current state and then
// reads the cluster status and writes it to current state.
func (etcd *Etcd) ReadCurrentState() error {
	// Copy previous state to existing state, so we can refresh it
	etcd.currentState = etcd.PreviousState

	return etcd.currentState.Read()
}

// Plan compares current state with desired state and sets cluster steps
// which needs to be applied.
func (etcd *Etcd) Plan() error {
	var steps step.Steps
	if etcd.currentState == nil {
		return fmt.Errorf("can't plan without knowing current state of the cluster")
	}

	// Iterate over previous state to find nodes, which should be removed
	for i, node := range etcd.currentState.Nodes {
		if etcd.DesiredState.Nodes[i] == nil {
			step, err := step.RemoveNodeStep(node)
			if err != nil {
				return fmt.Errorf("failed to create RemoveNode step: %s", err)
			}
			steps = append(steps, step)
		}
	}

	// Iterate over desired state to find which nodes should be created
	for _, node := range etcd.DesiredState.Nodes {
		step, err := step.AddNodeStep(node)
		if err != nil {
			return fmt.Errorf("failed to create AddNode step: %s", err)
		}
		steps = append(steps, step)
	}

	etcd.Steps = steps
	return nil
}

// SetImage sets which image should be used on all nodes.
// If this is not called, image needs to be set on all nodes.
func (etcd *Etcd) SetImage(image string) error {
	// TODO add validation logic?
	etcd.DesiredState.Image = image

	return nil
}

// PresentPlan prints planned steps in human-friendly form.
func (etcd *Etcd) PresentPlan() {
	for i, step := range etcd.Steps {
		desc, err := step.Describe()
		if err != nil {
			fmt.Println(fmt.Sprintf("Unable to describe step %d: %s", i+1, err))
			continue
		}
		fmt.Println(fmt.Sprintf("%d: %s", i+1, desc))
	}
}
