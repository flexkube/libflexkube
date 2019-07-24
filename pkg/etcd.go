package etcd

// Etcd stores all etcd cluster states
type Etcd struct {
	// PreviousState can be loaded from any data store
	// It is required to have previous state to be able to remove old elements
	// This struct is public and can be interacted with from outside
	PreviousState *EtcdState
	// CurrentState is loaded based on the PreviousState by doing health checks
	// on each node etc.
	// This struct is private and can be modified only by Etcd functions
	currentState *EtcdState
	// DesiredState represents new state requested by user. Deployment process
	// checks what is missing in CurrentState and attempts to fullfil it
	// This struct is public and can be interacted with from outside
	DesiredState *EtcdState
}

func New() *Etcd {
	previousState := NewState()
	currentState := NewState()
	desiredState := NewState()

	return &Etcd{
		PreviousState: previousState,
		currentState:  currentState,
		DesiredState:  desiredState,
	}
}
