// Package types contains types used for managing the containers. They are put in separate package
// to avoid cyclic dependencies while importing.
package types

// ContainerConfig stores runtime-agnostic information how to run the container.
type ContainerConfig struct {
	// Name is a name of the container.
	Name string `json:"name"`

	// Image is a container image to use.
	Image string `json:"image"`

	// Args is a list of arguments to pass to the container.
	Args []string `json:"args,omitempty"`

	// Entrypoint is a binary, which will be started in the container.
	Entrypoint []string `json:"entrypoint,omitempty"`

	// Ports is a list of ports, which will be exposed by the container.
	Ports []PortMap `json:"ports,omitempty"`

	// Mounts is a list of mounts, which will be added to the container.
	Mounts []Mount `json:"mounts,omitempty"`

	// Privileged controls, if created container should have full access to the
	// host.
	Privileged bool `json:"privileged,omitempty"`

	// NetworkMode defines what network the container should use.
	//
	// Valid values depends on used container runtime.
	NetworkMode string `json:"networkMode,omitempty"`

	// PidMode defines, in which PID namespace container should run.
	//
	// Valid values depends on used container runtime.
	PidMode string `json:"pidMode,omitempty"`

	// IpcMode defines, in which IPC namespace container should run.
	//
	// Valid values depends on used container runtime.
	IpcMode string `json:"ipcMode,omitempty"`

	// User defines, as which user the container should run.
	User string `json:"user,omitempty"`

	// Group defines as which group the container should run.
	Group string `json:"group,omitempty"`

	// Env defines a key-value environment variables to set in the container.
	Env map[string]string `json:"env,omitempty"`
}

// ContainerStatus stores status information received from the runtime.
//
// TODO: This should cover all fields which are defined in ContainerConfig,
// so we can read and compare if actual configuration matches our expectations.
type ContainerStatus struct {
	// ID is a runtime specific container ID.
	ID string `json:"id,omitempty"`

	// Status is a runtime specific status string.
	Status string `json:"status,omitempty"`
}

// PortMap is basically a github.com/docker/go-connections/nat.PortMap.
//
// TODO: Once we introduce Kubelet runtime, we need to figure out how to structure it.
type PortMap struct {
	// IP is an IP address on which container port should be exposed.
	IP string `json:"ip"`

	// Port defines, which port should be exposed.
	Port int `json:"port"`

	// Protocol defines what protocol should be exposed from the container.
	Protocol string `json:"protocol"`
}

// Mount describe host bind mount.
//
// TODO: Same as PortMap.
type Mount struct {
	// Source is a host filesystem path which will be mounted into the container.
	//
	// Example value: '/opt/bin'.
	Source string `json:"source"`

	// Target is a path in container's filesystem where host path will be mounted.
	//
	// Example value: '/usr/local/bin'.
	Target string `json:"target"`

	// Propagation defines how the mounts in host path will be propatated.
	//
	// Valid value depends on used container runtime.
	Propagation string `json:"propagation,omitempty"`
}

// File describes file, which can be either copied to or from container.
type File struct {
	// Path is a path on the filesystem.
	Path string `json:"path"`

	// Content is a content of the file. Binary files are currently not supported.
	Content string `json:"content"`

	// Mode is a numeric file mode.
	Mode int64 `json:"mode"`

	// User is an owner of the file.
	User string `json:"uid"`

	// Group is a group owner of the file.
	Group string `json:"gid"`
}

// Exists controls, how container existence is determined based on ContainerStatus.
// If state has no ID set, it means that container does not exist.
func (s *ContainerStatus) Exists() bool {
	return s.ID != ""
}

// Running determines if container is running, based on ContainerStatus.
func (s *ContainerStatus) Running() bool {
	return s.Exists() && s.Status == "running"
}

// Restarting returns true, if container is restarting in a loop, based on ContainerStatus.
func (s *ContainerStatus) Restarting() bool {
	return s.Exists() && s.Status == "restarting"
}
