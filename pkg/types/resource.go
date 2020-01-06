package types

// Resource interface defines flexkube resource like kubelet pool or static controlplane.
type Resource interface {
	StateToYaml() ([]byte, error)
	CheckCurrentState() error
	Deploy() error
}
