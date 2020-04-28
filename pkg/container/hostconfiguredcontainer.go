package container

import (
	"fmt"
	"os"
	"path"

	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
)

// ResourceInstance interface represents struct, which can be converted to HostConfiguredContainer.
type ResourceInstance interface {
	ToHostConfiguredContainer() (*HostConfiguredContainer, error)
}

// HostConfiguredContainerInterface exports hostConfiguredContainer capabilities.
type HostConfiguredContainerInterface interface {
	ConfigurationStatus() error
	Configure(paths []string) error
	Create() error
	Status() error
	Start() error
	Stop() error
	Delete() error
}

const (
	// ConfigMountpoint is where host file-system is mounted in the -config container.
	ConfigMountpoint = "/mnt/host"

	// configFileMode is default configuration file permissions.
	configFileMode = 0600

	// mountpointDirMode is default host mountpoint directory permission.
	mountpointDirMode = 0755
)

// Hooks defines type of hooks HostConfiguredContainer supports.
type Hooks struct {
	PostStart *Hook
}

// Hook is an action, which may be called before or after certain container operation, like starting or creating.
type Hook func() error

// HostConfiguredContainer represents single container, running on remote host with it's configuration files.
type HostConfiguredContainer struct {
	Container   Container         `json:"container"`
	Host        host.Host         `json:"host"`
	ConfigFiles map[string]string `json:"configFiles,omitempty"`

	Hooks *Hooks `json:"-"`
}

// hostConfiguredContainer is a validated version of HostConfiguredContainer, which allows user to perform
// actions on it.
type hostConfiguredContainer struct {
	container       Interface
	host            host.Host
	configFiles     map[string]string
	configContainer InstanceInterface
	hooks           *Hooks
}

// New validates HostConfiguredContainer struct and return it's executable version.
func (m *HostConfiguredContainer) New() (HostConfiguredContainerInterface, error) {
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate container configuration: %w", err)
	}

	c, _ := m.Container.New()

	hcc := &hostConfiguredContainer{
		container:   c,
		host:        m.Host,
		configFiles: m.ConfigFiles,
		hooks:       m.Hooks,
	}

	if hcc.hooks == nil {
		hcc.hooks = &Hooks{}
	}

	return hcc, nil
}

// Validate validates HostConfiguredContainer struct. All validation rules should be placed here.
func (m *HostConfiguredContainer) Validate() error {
	if err := m.Container.Validate(); err != nil {
		return fmt.Errorf("failed to validate container configuration: %w", err)
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
	h, err := m.host.New()
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
	c := m.container.RuntimeConfig()

	// Store originally configured address so we can restore it later.
	a := c.GetAddress()

	s, err := m.connectAndForward(a)
	if err != nil {
		return fmt.Errorf("forwarding host failed: %w", err)
	}

	// Override configuration with forwarded address and create Runtime from it.
	c.SetAddress(s)

	r, err := c.New()
	if err != nil {
		return err
	}

	// Use forwarded Runtime for managing container.
	m.container.SetRuntime(r)

	// Restore original address in the runtime configuration (as nested forwarding won't work).
	c.SetAddress(a)

	ro, err := c.New()
	if err != nil {
		return err
	}

	// After we're done calling action, restore original runtime to the container.
	defer m.container.SetRuntime(ro)

	return action()
}

// createConfigurationContainer creates container used for reading and updating configuration and
// stores saves it reference.
func (m *hostConfiguredContainer) createConfigurationContainer() error {
	cc := &container{
		base: base{
			config: types.ContainerConfig{
				Name:  fmt.Sprintf("%s-config", m.container.Config().Name),
				Image: m.container.Config().Image,
				Mounts: []types.Mount{
					{
						Source: "/",
						Target: ConfigMountpoint,
					},
				},
			},
			runtime: m.container.Runtime(),
		},
	}

	// Docker container does not need to run (be started) to be able to copy files from it.
	// TODO: This might not be the case for other container runtimes.
	ci, err := cc.Create()
	if err != nil {
		return fmt.Errorf("failed creating config container while checking configuration: %w", err)
	}

	m.configContainer = ci

	return nil
}

// removeConfigurationContainer removes configuration container created with createConfigurationContainer.
// If container does not exist, nil is immediately returned, which makes this function idempotent.
func (m *hostConfiguredContainer) removeConfigurationContainer() error {
	s, err := m.configContainer.Status()
	if err != nil {
		return fmt.Errorf("failed checking if container exists: %w", err)
	}

	if s.ID == "" {
		return nil
	}

	return m.configContainer.Delete()
}

// updateConfigurationStatus overrides configFiles field with current content of configuration files.
// If configuration file is missing, the entry is removed from the map.
func (m *hostConfiguredContainer) updateConfigurationStatus() error {
	// If there is no config files configured, don't do anything.
	if len(m.configFiles) == 0 {
		return nil
	}

	// Build list of files we need to read from the container.
	files := []string{}

	// Keep map of original paths.
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

// Configure copies specified configuration files on target host.
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
			Mode:    configFileMode,
			User:    m.container.Config().User,
			Group:   m.container.Config().Group,
		})
	}

	if err := m.configContainer.Copy(files); err != nil {
		return err
	}

	return nil
}

// statMounts fetches information about mounts on the host.
func (m *hostConfiguredContainer) statMounts() (map[string]os.FileMode, error) {
	paths := []string{}

	// Loop over mount points.
	for _, m := range m.dirMounts() {
		paths = append(paths, path.Join(ConfigMountpoint, m.Source))
	}

	// Don't execute stat at all if there is no files to stat.
	if len(paths) == 0 {
		return map[string]os.FileMode{}, nil
	}

	return m.configContainer.Stat(paths)
}

// isDirMount checks if given path is intended to be a directory by checking for a
// trailing slash.
func (m *hostConfiguredContainer) dirMounts() []types.Mount {
	r := []types.Mount{}

	for _, m := range m.container.Config().Mounts {
		if m.Source[len(m.Source)-1:] == "/" {
			r = append(r, m)
		}
	}

	return r
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

	// Collect missing mountpoints.
	files := []*types.File{}

	for _, m := range m.dirMounts() {
		p := path.Join(ConfigMountpoint, m.Source)
		fm, exists := rc[p]

		// If path exists as a file, it can't be mounted as a directory, so fail.
		if exists && !fm.IsDir() {
			return fmt.Errorf("mountpoint %s exists as file", m.Source)
		}

		// If mountpoint does not exist, and it's name has a trailing slash, we should create it as a directory.
		if !exists {
			files = append(files, &types.File{
				Path: fmt.Sprintf("%s/", p),
				Mode: mountpointDirMode,
			})
		}
	}

	// If there is no mountpoints to create, don't call the runtime again.
	if len(files) == 0 {
		return nil
	}

	// Create missing mountpoints.
	return m.configContainer.Copy(files)
}

// Create creates new container on target host.
func (m *hostConfiguredContainer) Create() error {
	return m.withForwardedRuntime(func() error {
		return m.withConfigurationContainer(func() error {
			if err := m.createMissingMounts(); err != nil {
				return fmt.Errorf("failed creating missing mountpoints: %w", err)
			}

			i, err := m.container.Create()
			if err != nil {
				return fmt.Errorf("failed creating container: %w", err)
			}

			s, err := i.Status()
			if err != nil {
				return fmt.Errorf("failed getting container status: %w", err)
			}

			*m.container.Status() = s

			return nil
		})
	})
}

// Status updates container status.
func (m *hostConfiguredContainer) Status() error {
	// If container does not exist, skip checking the status of it, as it won't work.
	if !m.container.Status().Exists() {
		return nil
	}

	return m.withForwardedRuntime(m.container.UpdateStatus)
}

// Start starts created container.
func (m *hostConfiguredContainer) Start() error {
	return withHook(nil, func() error {
		return m.withForwardedRuntime(m.container.Start)
	}, m.hooks.PostStart)
}

// Stop stops created container.
func (m *hostConfiguredContainer) Stop() error {
	return m.withForwardedRuntime(m.container.Stop)
}

// Delete removes node's data and removes the container.
func (m *hostConfiguredContainer) Delete() error {
	return m.withForwardedRuntime(m.container.Delete)
}

// withHook wraps given action function with pre and post functionality.
//
// This allows to inject custom actions before and after hostConfiguredContainer operations.
func withHook(preHook *Hook, action func() error, postHook *Hook) error {
	if preHook != nil {
		if err := (*preHook)(); err != nil {
			return err
		}
	}

	if err := action(); err != nil {
		return err
	}

	if postHook == nil {
		return nil
	}

	return (*postHook)()
}
