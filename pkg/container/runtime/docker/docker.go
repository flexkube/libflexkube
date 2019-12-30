package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
)

// Config struct represents Docker container runtime configuration.
type Config struct {
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
}

// docker struct is a struct, which can be used to manage Docker containers.
type docker struct {
	ctx context.Context
	cli *client.Client
}

// SetAddress sets runtime config address where it should connect.
func (c *Config) SetAddress(s string) {
	c.Host = s
}

// GetAddress returns configured container runtime address.
func (c *Config) GetAddress() string {
	if c != nil && c.Host != "" {
		return c.Host
	}

	return client.DefaultDockerHost
}

// New validates Docker runtime configuration and returns configured
// runtime client.
func (c *Config) New() (runtime.Runtime, error) {
	opts := []client.Opt{
		client.WithVersion(defaults.DockerAPIVersion),
	}

	if c != nil && c.Host != "" {
		opts = append(opts, client.WithHost(c.Host))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("creating Docker client: %w", err)
	}

	return &docker{
		ctx: context.Background(),
		cli: cli,
	}, nil
}

// Start starts Docker container.
func (d *docker) Create(config *types.ContainerConfig) (string, error) {
	// Pull image to make sure it's available.
	// TODO make it configurable?
	out, err := d.cli.ImagePull(d.ctx, config.Image, dockertypes.ImagePullOptions{})
	if err != nil {
		return "", fmt.Errorf("pulling image: %w", err)
	}

	defer out.Close()

	if _, err := io.Copy(ioutil.Discard, out); err != nil {
		return "", fmt.Errorf("failed to pull image: %w", err)
	}

	// TODO That should be validated at ContainerConfig level!
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	for _, ip := range config.Ports {
		port, err := nat.NewPort(ip.Protocol, strconv.Itoa(ip.Port))
		if err != nil {
			return "", fmt.Errorf("failed mapping ports: %w", err)
		}

		if _, exists := portBindings[port]; !exists {
			portBindings[port] = []nat.PortBinding{}
		}

		portBindings[port] = append(portBindings[port], nat.PortBinding{
			HostIP:   ip.IP,
			HostPort: strconv.Itoa(ip.Port),
		})
		exposedPorts[port] = struct{}{}
	}

	// TODO validate that
	mounts := []mount.Mount{}
	for _, m := range config.Mounts {
		mounts = append(mounts, mount.Mount{
			Type:   "bind",
			Source: m.Source,
			Target: m.Target,
			// TODO validate!
			BindOptions: &mount.BindOptions{
				Propagation: mount.Propagation(m.Propagation),
			},
		})
	}

	// Just structs required for starting container.
	dockerConfig := containertypes.Config{
		Image:        config.Image,
		Cmd:          config.Args,
		Entrypoint:   config.Entrypoint,
		ExposedPorts: exposedPorts,
	}
	hostConfig := containertypes.HostConfig{
		Mounts:       mounts,
		PortBindings: portBindings,
		Privileged:   config.Privileged,
		NetworkMode:  containertypes.NetworkMode(config.NetworkMode),
		PidMode:      containertypes.PidMode(config.PidMode),
		IpcMode:      containertypes.IpcMode(config.IpcMode),
		RestartPolicy: containertypes.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// Create container.
	c, err := d.cli.ContainerCreate(d.ctx, &dockerConfig, &hostConfig, &network.NetworkingConfig{}, config.Name)
	if err != nil {
		return "", fmt.Errorf("creating container: %w", err)
	}

	return c.ID, nil
}

// Start starts Docker container.
func (d *docker) Start(id string) error {
	return d.cli.ContainerStart(d.ctx, id, dockertypes.ContainerStartOptions{})
}

// Stop stops Docker container.
func (d *docker) Stop(id string) error {
	// TODO make this configurable?
	timeout := time.Duration(30) * time.Second
	return d.cli.ContainerStop(d.ctx, id, &timeout)
}

// Status returns container status.
func (d *docker) Status(id string) (*types.ContainerStatus, error) {
	status, err := d.cli.ContainerInspect(d.ctx, id)
	if err != nil {
		// If container is missing, return no status.
		if client.IsErrNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("inspecting container failed: %w", err)
	}

	return &types.ContainerStatus{
		Image:  status.Image,
		ID:     id,
		Name:   status.Name,
		Status: status.State.Status,
	}, nil
}

// Delete removes the container.
func (d *docker) Delete(id string) error {
	return d.cli.ContainerRemove(d.ctx, id, dockertypes.ContainerRemoveOptions{})
}

// Copy takes map of files and their content and copies it to the container using TAR archive.
//
// TODO Add support for base64 encoded content to support copying binary files.
func (d *docker) Copy(id string, files []*types.File) error {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	for _, f := range files {
		h := &tar.Header{
			Name: f.Path,
			Mode: f.Mode,
			Size: int64(len(f.Content)),
		}

		if err := tw.WriteHeader(h); err != nil {
			return err
		}

		if _, err := tw.Write([]byte(f.Content)); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}

	return d.cli.CopyToContainer(d.ctx, id, "/", buf, dockertypes.CopyToContainerOptions{})
}

// Stat check if given paths exist on the container.
func (d *docker) Stat(id string, paths []string) (map[string]*os.FileMode, error) {
	result := map[string]*os.FileMode{}

	for _, p := range paths {
		s, err := d.cli.ContainerStatPath(d.ctx, id, p)
		if err != nil && !client.IsErrNotFound(err) {
			return nil, err
		}

		if s.Name != "" {
			result[p] = &s.Mode
		}
	}

	return result, nil
}

// Read reads files from container.
func (d *docker) Read(id string, srcPaths []string) ([]*types.File, error) {
	files := []*types.File{}

	for _, f := range srcPaths {
		rc, ps, err := d.cli.CopyFromContainer(d.ctx, id, f)
		if err != nil && !client.IsErrNotFound(err) {
			return nil, fmt.Errorf("failed copying from container: %w", err)
		}

		// File does not exist.
		if ps.Name == "" {
			continue
		}

		buf := new(bytes.Buffer)
		tr := tar.NewReader(rc)

		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed unpacking tar header from Docker file copy: %w", err)
			}

			if header.Typeflag != tar.TypeReg {
				continue
			}

			if _, err := buf.ReadFrom(tr); err != nil {
				return nil, fmt.Errorf("failed reading from tar archive: %w", err)
			}
		}
		rc.Close()

		files = append(files, &types.File{
			Path:    f,
			Content: buf.String(),
			Mode:    int64(ps.Mode),
		})
	}

	return files, nil
}

// imageID lists images which are pulled on the host and looks for the tag given by the user.
//
// If image with given tag is found, it's ID is returned.
// If image is not pulled, empty string is returned.
//
// This method allows to check if the image is present on the host.
func (d *docker) imageID(image string) (string, error) {
	images, err := d.cli.ImageList(d.ctx, dockertypes.ImageListOptions{})
	if err != nil {
		return "", fmt.Errorf("listing docker images failed: %w", err)
	}

	for _, i := range images {
		for _, tag := range i.RepoTags {
			if tag == image {
				return i.ID, nil
			}
		}
	}

	return "", nil
}

// pullImage pulls specified container image.
func (d *docker) pullImage(image string) error {
	out, err := d.cli.ImagePull(d.ctx, image, dockertypes.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("pulling image failed: %w", err)
	}

	defer out.Close()

	if _, err := io.Copy(ioutil.Discard, out); err != nil {
		return fmt.Errorf("failed to discard pulling messages: %w", err)
	}

	return nil
}
