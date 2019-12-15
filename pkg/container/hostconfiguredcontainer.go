package container

import (
	"fmt"
	"os"
	"path"

	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
)

// ConfigMountpoint is where host file-system is mounted in the -config container.
const ConfigMountpoint = "/mnt/host"

// HostConfiguredContainer represents single container, running on remote host with it's configuration files
type HostConfiguredContainer struct {
	Container   Container         `json:"container" yaml:"container"`
	Host        host.Host         `json:"host" yaml:"host"`
	ConfigFiles map[string]string `json:"configFiles,omitempty" yaml:"configFiles,omitempty"`
}

// hostConfiguredContainer is a validated version of HostConfiguredContainer, which allows user to perform
// actions on it
type hostConfiguredContainer struct {
	container       Container
	host            host.Host
	configFiles     map[string]string
	configContainer *Container
}

// New validates HostConfiguredContainer struct and return it's executable version
func (m *HostConfiguredContainer) New() (*hostConfiguredContainer, error) {
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("failed to valide container configuration: %w", err)
	}

	return &hostConfiguredContainer{
		container:   m.Container,
		host:        m.Host,
		configFiles: m.ConfigFiles,
	}, nil
}

// Validate validates HostConfiguredContainer struct. All validation rules should be placed here.
func (m *HostConfiguredContainer) Validate() error {
	if err := m.Container.Validate(); err != nil {
		return fmt.Errorf("failed to valide container configuration: %w", err)
	}

	if err := m.Host.Validate(); err != nil {
		return fmt.Errorf("failed to validate host configuration: %w", err)
	}

	return nil
}

// connectAndForward instantiates new host object, connects to it and then
// forwards given UNIX socket using this connection.
//
// It returns address of local UNIX socket, where user can connect.
func (m *hostConfiguredContainer) connectAndForward(a string) (string, error) {
	h, err := host.New(&m.host)
	if err != nil {
		return "", err
	}

	hc, err := h.Connect()
	if err != nil {
		return "", err
	}

	s, err := hc.ForwardUnixSocket(a)
	if err != nil {
		return "", err
	}

	return s, nil
}

// withForwardedRuntime takes action function as an argument and before executing it, it configures the runtime
// address to be forwarded using SSH. After the action is finished, it restores original address of the runtime.
func (m *hostConfiguredContainer) withForwardedRuntime(action func() error) error {
	// Store originally configured address so we can restore it later
	a := m.container.Runtime.Docker.GetAddress()

	s, err := m.connectAndForward(a)
	if err != nil {
		return fmt.Errorf("forwarding host failed: %w", err)
	}

	m.container.Runtime.Docker.SetAddress(s)

	defer m.container.Runtime.Docker.SetAddress(a)

	return action()
}

// createConfigurationContainer creates container used for reading and updating configuration and
// stores saves it reference.
func (m *hostConfiguredContainer) createConfigurationContainer() error {
	c := &Container{
		Config: types.ContainerConfig{
			Name:  fmt.Sprintf("%s-config", m.container.Config.Name),
			Image: m.container.Config.Image,
			Mounts: []types.Mount{
				{
					Source: "/",
					Target: ConfigMountpoint,
				},
			},
		},
		Runtime: m.container.Runtime,
	}

	// Docker container does not need to run (be started) to be able to copy files from it
	// TODO this might not be the case for other container runtimes
	if err := c.Create(); err != nil {
		return fmt.Errorf("failed creating config container while checking configuration: %w", err)
	}

	m.configContainer = c

	return nil
}

// removeConfigurationContainer removes configuration container created with createConfigurationContainer.
// If container does not exist, nil is immediately returned, which makes this function idempotent.
func (m *hostConfiguredContainer) removeConfigurationContainer() error {
	if m.configContainer.Status == nil {
		return nil
	}

	return m.configContainer.Delete()
}

// updateConfigurationStatus overrides configFiles field with current content of configuration files.
// If configuration file is missing, the entry is removed from the map.
func (m *hostConfiguredContainer) updateConfigurationStatus() error {
	// Build list of files we need to read from the container
	files := []string{}

	// Keep map of original paths
	paths := map[string]string{}

	// Build list of the files we should read.
	for p := range m.configFiles {
		cpath := path.Join(ConfigMountpoint, p)
		files = append(files, cpath)
		paths[cpath] = p
	}

	f, err := m.configContainer.Read(files)
	if err != nil {
		return fmt.Errorf("failed to read configuration status: %w", err)
	}

	m.configFiles = map[string]string{}

	for _, f := range f {
		m.configFiles[paths[f.Path]] = f.Content
	}

	return nil
}

// withConfigurationContainer is a wrapper function for functions, which require functional
// configuration container reference. This function creates configuration container before executing
// desired action and makes sure it's removed after the action is finished.
//
// If error occurs in the desired action, this error is returned and configuration container is opportunistically
// removed. If that operation fails as well, error is only logged.
func (m *hostConfiguredContainer) withConfigurationContainer(action func() error) error {
	if err := m.createConfigurationContainer(); err != nil {
		return fmt.Errorf("failed to create container for managing configuration: %w", err)
	}

	defer func() {
		if err := m.removeConfigurationContainer(); err != nil {
			fmt.Printf("Removing configuration container failed: %v", err)
		}
	}()

	if err := action(); err != nil {
		return err
	}

	return m.removeConfigurationContainer()
}

// ConfigurationStatus updates configuration file struct with current state on the target host.
func (m *hostConfiguredContainer) ConfigurationStatus() error {
	return m.withForwardedRuntime(func() error {
		return m.withConfigurationContainer(m.updateConfigurationStatus)
	})
}

// Configure copies specified configuration files on target host
//
// It uses host definition to connect to container runtime, which is then used
// to create temporary container used for copying files and also bypassing privileges requirements.
//
// With Kubelet runtime, 'tar' binary is required on the container to be able to write and read the configurations.
// By default, the image which will be deployed will be used for copying the configuration as well, to avoid pulling
// multiple images, which will save disk space and time. If it happens that this image does not have 'tar' binary,
// user can override ConfigImage field in the configuration, to specify different image which should be
// pulled and used for configuration management.
func (m *hostConfiguredContainer) Configure(paths []string) error {
	return m.withForwardedRuntime(func() error {
		return m.withConfigurationContainer(func() error {
			return m.copyConfigFiles(paths)
		})
	})
}

// copyConfigFiles takes list of configuration files which should be created in the container
// and creates them in batch. This function requires functional config container.
func (m *hostConfiguredContainer) copyConfigFiles(paths []string) error {
	files := []*types.File{}

	for _, p := range paths {
		content, exists := m.configFiles[p]
		if !exists {
			return fmt.Errorf("can't configure file which do not exist: %s", p)
		}

		files = append(files, &types.File{
			Path:    path.Join(ConfigMountpoint, p),
			Content: content,
			Mode:    0600,
		})
	}

	if err := m.configContainer.Copy(files); err != nil {
		return err
	}

	return nil
}

// statMounts fetches information about mounts on the host.
func (m *hostConfiguredContainer) statMounts() (map[string]*os.FileMode, error) {
	paths := []string{}

	// Loop over mount points
	for _, m := range m.container.Config.Mounts {
		paths = append(paths, path.Join(ConfigMountpoint, m.Source))
	}

	return m.configContainer.Stat(paths)
}

// createMissingMounts creates missing host directories, which are requested for container.
//
// Requested mount source must have trailing slash ('/') in the name to be created as a directory.
// If requested directory mount is found on host file system as a file, the error is returned.
func (m *hostConfiguredContainer) createMissingMounts() error {
	// Get information about existing mountpoints.
	rc, err := m.statMounts()
	if err != nil {
		return fmt.Errorf("failed checking if mountpoints exist: %w", err)
	}

	// Collect missing mountpoints
	files := []*types.File{}

	for _, m := range m.container.Config.Mounts {
		p := path.Join(ConfigMountpoint, m.Source)
		fm, exists := rc[p]

		if exists && *fm == os.ModeDir {
			return fmt.Errorf("mountpoint %s exists as file", m.Source)
		}

		// If mountpoint does not exist, and it's name has a trailing slash, we should create it as a directory.
		if !exists && m.Source[len(m.Source)-1:] == "/" {
			files = append(files, &types.File{
				Path: fmt.Sprintf("%s/", p),
				Mode: 0755,
			})
		}
	}

	// Create missing mountpoints.
	if err := m.configContainer.Copy(files); err != nil {
		return fmt.Errorf("creating host mountpoints failed: %w", err)
	}

	return nil
}

// Create creates new container on target host.
func (m *hostConfiguredContainer) Create() error {
	return m.withForwardedRuntime(func() error {
		return m.withConfigurationContainer(func() error {
			if err := m.createMissingMounts(); err != nil {
				return fmt.Errorf("failed creating missing mountpoints: %w", err)
			}

			return m.container.Create()
		})
	})
}

// Status updates container status.
func (m *hostConfiguredContainer) Status() error {
	// If container does not exist, skip checking the status of it, as it won't work
	if m.container.Status == nil {
		return nil
	}

	return m.withForwardedRuntime(m.container.UpdateStatus)
}

// Start starts created container.
func (m *hostConfiguredContainer) Start() error {
	return m.withForwardedRuntime(m.container.Start)
}

// Stop stops created container.
func (m *hostConfiguredContainer) Stop() error {
	return m.withForwardedRuntime(m.container.Stop)
}

// Delete removes node's data and removes the container.
func (m *hostConfiguredContainer) Delete() error {
	return m.withForwardedRuntime(m.container.Delete)
}
