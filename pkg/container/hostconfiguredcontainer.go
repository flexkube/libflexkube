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
	container   Container
	host        host.Host
	configFiles map[string]string
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
// forwards connection to container runtime and reconfigures container runtime
// to connect to forwarded endpoint.
//
// TODO maybe we make this take a function to remove boilerplate from helper functions?
func (m *hostConfiguredContainer) connectAndForward() error {
	h, err := host.New(&m.host)
	if err != nil {
		return err
	}

	hc, err := h.Connect()
	if err != nil {
		return err
	}

	// TODO don't use docker directly
	a := m.container.Runtime.Docker.GetAddress()

	s, err := hc.ForwardUnixSocket(a)
	if err != nil {
		return err
	}

	m.container.Runtime.Docker.SetAddress(s)

	return nil
}

// createConfigurationContainer creates container used for reading and updating configuration.
//
// It returns original docker runtime address and created container.
// It is up to the user to later remove the container and restore original address of runtime
//
// TODO maybe configuration container should get it's own simple struct with methods?
func (m *hostConfiguredContainer) createConfigurationContainer() (string, *Container, error) {
	// Store originally configured address so we can restore it later
	a := m.container.Runtime.Docker.GetAddress()

	if err := m.connectAndForward(); err != nil {
		return "", nil, fmt.Errorf("forwarding host failed: %w", err)
	}

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
		return "", nil, fmt.Errorf("failed creating config container while checking configuration: %w", err)
	}

	return a, c, nil
}

// removeConfigurationContainer removes configuration container created with createConfigurationContainer
func (m *hostConfiguredContainer) removeConfigurationContainer(originalAddress string, c *Container) error {
	// Restore original address, so we don't get random values when we serialize back the object to store it
	// in the state.
	defer m.container.Runtime.Docker.SetAddress(originalAddress)

	// If container does not exist anymore, simply return without error. This makes this function idempotent.
	if c.Status == nil {
		return nil
	}

	return c.Delete()
}

// ConfigurationStatus updates configuration status
func (m *hostConfiguredContainer) ConfigurationStatus() error {
	a, c, err := m.createConfigurationContainer()
	if err != nil {
		return fmt.Errorf("failed to create container for managing configuration: %w", err)
	}
	defer m.removeConfigurationContainer(a, c)

	files := []string{}

	// Build list of the files we should read.
	for p := range m.configFiles {
		files = append(files, path.Join(ConfigMountpoint, p))
	}

	f, err := c.Read(files)
	if err != nil {
		return fmt.Errorf("failed to read configuration status: %w", err)
	}

	m.configFiles = map[string]string{}

	for _, f := range f {
		m.configFiles[f.Path] = f.Content
	}

	return m.removeConfigurationContainer(a, c)
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
	a, c, err := m.createConfigurationContainer()
	if err != nil {
		return fmt.Errorf("failed to create container for managing configuration: %w", err)
	}

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

	if err := c.Copy(files); err != nil {
		return err
	}

	return m.removeConfigurationContainer(a, c)
}

// Create creates new container on target host
func (m *hostConfiguredContainer) Create() error {
	// Before we create a container, make sure all mounts exists on the host using config container
	a, c, err := m.createConfigurationContainer()
	if err != nil {
		return fmt.Errorf("failed to create container for managing configuration: %w", err)
	}
	defer m.removeConfigurationContainer(a, c)

	files := []*types.File{}

	// Loop over mount points
	for _, m := range m.container.Config.Mounts {
		fp := path.Join(ConfigMountpoint, m.Source)

		rc, err := c.Stat(fp)
		if err != nil {
			return fmt.Errorf("failed reading file %s: %w", m.Source, err)
		}

		if rc != nil && *rc == os.ModeDir {
			return fmt.Errorf("mountpoint %s exists as file", m.Source)
		}

		// TODO perhaps path handling should be improved here
		if rc == nil && m.Source[len(m.Source)-1:] == "/" {
			files = append(files, &types.File{
				Path: fmt.Sprintf("%s/", path.Join(ConfigMountpoint, m.Source)),
				Mode: 0755,
			})
		}
	}

	if err := c.Copy(files); err != nil {
		return fmt.Errorf("creating host mountpoints failed: %w", err)
	}

	if err := m.container.Create(); err != nil {
		return fmt.Errorf("creating failed: %w", err)
	}

	return m.removeConfigurationContainer(a, c)
}

// Status updates container status
func (m *hostConfiguredContainer) Status() error {
	// If container does not exist, skip checking the status of it, as it won't work
	if m.container.Status == nil {
		return nil
	}

	// TODO maybe we can cache forwarding somehow?
	a := m.container.Runtime.Docker.GetAddress()

	if err := m.connectAndForward(); err != nil {
		return fmt.Errorf("forwarding host failed: %w", err)
	}

	if err := m.container.UpdateStatus(); err != nil {
		return fmt.Errorf("updating status failed: %w", err)
	}

	m.container.Runtime.Docker.SetAddress(a)

	return nil
}

// Start starts created container
// TODO plenty of boilerplate code here, maybe create executeForwarded method
// which takes function as an argument to clean it up?
func (m *hostConfiguredContainer) Start() error {
	a := m.container.Runtime.Docker.GetAddress()

	if err := m.connectAndForward(); err != nil {
		return fmt.Errorf("forwarding host failed: %w", err)
	}

	if err := m.container.Start(); err != nil {
		return fmt.Errorf("starting failed: %w", err)
	}

	m.container.Runtime.Docker.SetAddress(a)

	return nil
}

// Stop stops created container
func (m *hostConfiguredContainer) Stop() error {
	a := m.container.Runtime.Docker.GetAddress()

	if err := m.connectAndForward(); err != nil {
		return fmt.Errorf("forwarding host failed: %w", err)
	}

	if err := m.container.Stop(); err != nil {
		return fmt.Errorf("stopping failed: %w", err)
	}

	m.container.Runtime.Docker.SetAddress(a)

	return nil
}

// Delete removes node's data and removes the container
func (m *hostConfiguredContainer) Delete() error {
	a := m.container.Runtime.Docker.GetAddress()

	if err := m.connectAndForward(); err != nil {
		return fmt.Errorf("forwarding host failed: %w", err)
	}

	if err := m.container.Delete(); err != nil {
		return fmt.Errorf("creating failed: %w", err)
	}

	m.container.Runtime.Docker.SetAddress(a)

	return nil
}
