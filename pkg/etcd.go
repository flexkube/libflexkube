package etcd

import "fmt"

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
}

func New() *Etcd {
	previousState := NewState()
	desiredState := NewState()

	return &Etcd{
		PreviousState: previousState,
		DesiredState:  desiredState,
	}
}

func (etcd *Etcd) AddNode(node *Node) error {
	return etcd.DesiredState.AddNode(node)
}

func (etcd *Etcd) LoadPreviousState() error {
	return nil
}

func (etcd *Etcd) ReadCurrentState() error {
	// Copy previous state to existing state, so we can refresh it
	etcd.currentState = etcd.PreviousState

	return etcd.currentState.Read()
}

func (etcd *Etcd) Plan() error {
	if etcd.currentState == nil {
		return fmt.Errorf("can't plan without knowing current state of the cluster")
	}
	for i, node := range etcd.DesiredState.Nodes {
		if node == nil {
			fmt.Println(fmt.Sprintf("Node '%s' should be created.", i))
		}
	}
	return nil
}

func (etcd *Etcd) SetImage(image string) error {
	// TODO add validation logic?
	etcd.DesiredState.Image = image

	return nil
}
