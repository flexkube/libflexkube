// Package types contains types used for managing the containers. They are put in separate package
// to avoid cyclic dependencies while importing.
package types

// ContainerConfig stores runtime-agnostic information how to run the container.
type ContainerConfig struct {
	Name        string    `json:"name"`
	Image       string    `json:"image"`
	Args        []string  `json:"args,omitempty"`
	Entrypoint  []string  `json:"entrypoint,omitempty"`
	Ports       []PortMap `json:"ports,omitempty"`
	Mounts      []Mount   `json:"mounts,omitempty"`
	Privileged  bool      `json:"privileged,omitempty"`
	NetworkMode string    `json:"networkMode,omitempty"`
	PidMode     string    `json:"pidMode,omitempty"`
	IpcMode     string    `json:"ipcMode,omitempty"`
	User        string    `json:"user,omitempty"`
	Group       string    `json:"group,omitempty"`
}

// ContainerStatus stores status information received from the runtime
// TODO this should cover all fields which are defined in ContainerConfig,
// so we can read and compare if actual configuration matches our expectations.
type ContainerStatus struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
}

// PortMap is basically a github.com/docker/go-connections/nat.PortMap.
//
// TODO: Once we introduce Kubelet runtime, we need to figure out how to structure it.
type PortMap struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// Mount describe host bind mount.
//
// TODO: Same as PortMap.
type Mount struct {
	Source      string `json:"source"`
	Target      string `json:"target"`
	Propagation string `json:"propagation,omitempty"`
}

// File describes file, which can be either copied to or from container.
type File struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Mode    int64  `json:"mode"`
	User    string `json:"uid"`
	Group   string `json:"gid"`
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
