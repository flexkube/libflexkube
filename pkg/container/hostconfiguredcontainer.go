package container

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/host"
)

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
				types.Mount{
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
	if err := c.Delete(); err != nil {
		return fmt.Errorf("failed removing config container while checking configuration: %w", err)
	}

	// Restore original address, so we don't get random values when we serialize back the object to store it
	// in the state.
	m.container.Runtime.Docker.SetAddress(originalAddress)

	return nil
}

// ConfigurationStatus updates configuration status
func (m *hostConfiguredContainer) ConfigurationStatus() error {
	a, c, err := m.createConfigurationContainer()
	if err != nil {
		return fmt.Errorf("failed to create container for managing configuration: %w", err)
	}

	// Loop over defined config files, check if they are on the host and update configFiles
	for p := range m.configFiles {
		fp := path.Join(ConfigMountpoint, p)
		// TODO kubelet support batch copying. Read interface should probably do that too and then
		// docker interface (which doesn't support batching) should simply imitate it.
		rc, err := c.Read(fp)
		if err != nil {
			return fmt.Errorf("failed reading file %s: %w", p, err)
		}
		// If ReadCloser is nil, this means that file does not exist in container
		// so we should remove the entry from the config map and move to next file.
		if rc == nil {
			delete(m.configFiles, p)
			continue
		}

		// If file exists, we read it and update content in the config map.
		if rc != nil {
			tr := tar.NewReader(rc)
			for {
				header, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}
				if header.Typeflag == tar.TypeReg {
					buf := new(bytes.Buffer)
					_, err := buf.ReadFrom(tr)
					if err != nil {
						return err
					}
					m.configFiles[p] = buf.String()
				}
			}
			rc.Close()
		}
	}

	return m.removeConfigurationContainer(a, c)
}

// Configure copies specified configuration file on target host
//
// It uses host definition to connect to container runtime, which is then used
// to create temporary container used for copying files and also bypassing privileges requirements.
//
// With Kubelet runtime, 'tar' binary is required on the container to be able to write and read the configurations.
// By default, the image which will be deployed will be used for copying the configuration as well, to avoid pulling
// multiple images, which will save disk space and time. If it happens that this image does not have 'tar' binary,
// user can override ConfigImage field in the configuration, to specify different image which should be
// pulled and used for configuration management.
func (m *hostConfiguredContainer) Configure(p string) error {
	a, c, err := m.createConfigurationContainer()
	if err != nil {
		return fmt.Errorf("failed to create container for managing configuration: %w", err)
	}

	content, exists := m.configFiles[p]
	if !exists {
		return fmt.Errorf("can't configure file which do not exist: %s", p)
	}

	buff := new(bytes.Buffer)
	tw := tar.NewWriter(buff)
	h := &tar.Header{
		Name: p,
		Mode: 0600,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(h); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := c.Copy(ConfigMountpoint, buff); err != nil {
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

	// Loop over mount points
	for _, m := range m.container.Config.Mounts {
		fp := path.Join(ConfigMountpoint, m.Source)
		rc, err := c.Stat(fp)
		if err != nil {
			return fmt.Errorf("failed reading file %s: %w", m.Source, err)
		}
		if rc != nil && *rc == os.ModeDir {
			return fmt.Errorf("mountpoint %s exists as file!", m.Source)
		}
		// TODO perhaps path handling should be improved here
		if rc == nil && m.Source[len(m.Source)-1:] == "/" {
			buff := new(bytes.Buffer)
			tw := tar.NewWriter(buff)
			h := &tar.Header{
				Name: fmt.Sprintf("%s/", m.Source),
				Mode: 0755,
			}
			if err := tw.WriteHeader(h); err != nil {
				return err
			}
			if err := tw.Close(); err != nil {
				return err
			}
			if err := c.Copy(ConfigMountpoint, buff); err != nil {
				return err
			}
		}
	}

	if err := m.container.Create(); err != nil {
		return fmt.Errorf("creating failed: %w", err)
	}
	return m.removeConfigurationContainer(a, c)
}

// Status updates container status
func (m *hostConfiguredContainer) Status() error {
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
