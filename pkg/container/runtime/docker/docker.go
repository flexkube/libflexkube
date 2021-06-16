// Package docker implements runtime.Interface and runtime.Config interfaces
// by talking to Docker API.
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
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/flexkube/libflexkube/internal/util"
	"github.com/flexkube/libflexkube/pkg/container/runtime"
	"github.com/flexkube/libflexkube/pkg/container/types"
	"github.com/flexkube/libflexkube/pkg/defaults"
)

const (
	// stopTimeout is how long we wait when gracefully stopping the container before force-killing it.
	stopTimeout = 30 * time.Second
)

// Config struct represents Docker container runtime configuration.
type Config struct {
	// Host is a Docker runtime URL. Usually 'unix:///run/docker.sock'. If empty
	// Docker's default URL will be used.
	Host string `json:"host,omitempty"`
}

// dockerClient is a wrapper interface over
// https://godoc.org/github.com/docker/docker/client#ContainerAPIClient
// with the functions we use.
type dockerClient interface { //nolint:dupl
	ContainerCreate(ctx context.Context, config *containertypes.Config, hostConfig *containertypes.HostConfig, networkingConfig *networktypes.NetworkingConfig, platform *v1.Platform, containerName string) (containertypes.ContainerCreateCreatedBody, error)
	ContainerStart(ctx context.Context, container string, options dockertypes.ContainerStartOptions) error
	ContainerStop(ctx context.Context, container string, timeout *time.Duration) error
	ContainerInspect(ctx context.Context, container string) (dockertypes.ContainerJSON, error)
	ContainerRemove(ctx context.Context, container string, options dockertypes.ContainerRemoveOptions) error
	CopyFromContainer(ctx context.Context, container, srcPath string) (io.ReadCloser, dockertypes.ContainerPathStat, error)
	CopyToContainer(ctx context.Context, container, path string, content io.Reader, options dockertypes.CopyToContainerOptions) error
	ContainerStatPath(ctx context.Context, container, path string) (dockertypes.ContainerPathStat, error)
	ImageList(ctx context.Context, options dockertypes.ImageListOptions) ([]dockertypes.ImageSummary, error)
	ImagePull(ctx context.Context, ref string, options dockertypes.ImagePullOptions) (io.ReadCloser, error)
}

// docker struct is a struct, which can be used to manage Docker containers.
type docker struct {
	ctx context.Context
	cli dockerClient
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
	cli, err := c.getDockerClient()
	if err != nil {
		return nil, fmt.Errorf("creating Docker client: %w", err)
	}

	return &docker{
		ctx: context.Background(),
		cli: cli,
	}, nil
}

func (c *Config) getDockerClient() (*client.Client, error) {
	opts := []client.Opt{
		client.WithVersion(defaults.DockerAPIVersion),
	}

	if c != nil && c.Host != "" {
		opts = append(opts, client.WithHost(c.Host))
	}

	return client.NewClientWithOpts(opts...)
}

// pullImageIfNotPresent pulls image if it's not already present on the host.
func (d *docker) pullImageIfNotPresent(image string) error {
	// Pull image to make sure it's available.
	// TODO make it configurable?
	id, err := d.imageID(image)
	if err != nil {
		return fmt.Errorf("failed checking for image presence: %w", err)
	}

	if id != "" {
		return nil
	}

	return d.pullImage(image)
}

// buildPorts converts container PortMap type to Docker port maps.
func buildPorts(ports []types.PortMap) (nat.PortMap, nat.PortSet, error) {
	// TODO That should be validated at ContainerConfig level!
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	for _, ip := range ports {
		port, err := nat.NewPort(ip.Protocol, strconv.Itoa(ip.Port))
		if err != nil {
			return nil, nil, fmt.Errorf("failed mapping ports: %w", err)
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

	return portBindings, exposedPorts, nil
}

// mounts converts container Mount to Docker mount type.
func mounts(m []types.Mount) []mount.Mount {
	mounts := []mount.Mount{}

	for _, m := range m {
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

	return mounts
}

func convertContainerConfig(config *types.ContainerConfig) (*containertypes.Config, *containertypes.HostConfig, error) {
	// TODO That should be validated at ContainerConfig level!
	portBindings, exposedPorts, err := buildPorts(config.Ports)
	if err != nil {
		return nil, nil, fmt.Errorf("failed building ports: %w", err)
	}

	u := config.User
	if config.Group != "" {
		u = fmt.Sprintf("%s:%s", config.User, config.Group)
	}

	env := []string{}
	for k, v := range config.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Just structs required for starting container.
	dockerConfig := containertypes.Config{
		Image:        config.Image,
		Cmd:          config.Args,
		Entrypoint:   config.Entrypoint,
		ExposedPorts: exposedPorts,
		User:         u,
		Env:          env,
	}
	hostConfig := containertypes.HostConfig{
		Mounts:       mounts(config.Mounts),
		PortBindings: portBindings,
		Privileged:   config.Privileged,
		NetworkMode:  containertypes.NetworkMode(config.NetworkMode),
		PidMode:      containertypes.PidMode(config.PidMode),
		IpcMode:      containertypes.IpcMode(config.IpcMode),
		RestartPolicy: containertypes.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	return &dockerConfig, &hostConfig, nil
}

// Start starts Docker container.
func (d *docker) Create(config *types.ContainerConfig) (string, error) {
	if err := d.pullImageIfNotPresent(config.Image); err != nil {
		return "", fmt.Errorf("failed pulling image: %w", err)
	}

	dockerConfig, hostConfig, err := convertContainerConfig(config)
	if err != nil {
		return "", fmt.Errorf("converting container config to Docker configuration: %w", err)
	}

	// Create container.
	c, err := d.cli.ContainerCreate(d.ctx, dockerConfig, hostConfig, &networktypes.NetworkingConfig{}, nil, config.Name)
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
	// TODO make timeout configurable?
	timeout := stopTimeout

	return d.cli.ContainerStop(d.ctx, id, &timeout)
}

// Status returns container status.
func (d *docker) Status(id string) (types.ContainerStatus, error) {
	s := types.ContainerStatus{
		ID: id,
	}

	status, err := d.cli.ContainerInspect(d.ctx, id)
	if err != nil {
		// If container is missing, return status with empty ID.
		if client.IsErrNotFound(err) {
			s.ID = ""

			return s, nil
		}

		return s, fmt.Errorf("inspecting container failed: %w", err)
	}

	s.Status = status.State.Status

	return s, nil
}

// Delete removes the container.
func (d *docker) Delete(id string) error {
	return d.cli.ContainerRemove(d.ctx, id, dockertypes.ContainerRemoveOptions{})
}

// Copy takes map of files and their content and copies it to the container using TAR archive.
//
// TODO Add support for base64 encoded content to support copying binary files.
func (d *docker) Copy(id string, files []*types.File) error {
	t, err := filesToTar(files)
	if err != nil {
		return fmt.Errorf("failed packing files to TAR archive: %w", err)
	}

	return d.cli.CopyToContainer(d.ctx, id, "/", t, dockertypes.CopyToContainerOptions{})
}

// filesToTar converts list of container files to tar archive format.
func filesToTar(files []*types.File) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	for _, f := range files {
		h := &tar.Header{
			Name:    f.Path,
			Mode:    f.Mode,
			Size:    int64(len(f.Content)),
			ModTime: time.Now(),
			Uname:   f.User,
			Gname:   f.Group,
		}

		if uid, err := strconv.Atoi(f.User); err == nil {
			h.Uid = uid
		}

		if gid, err := strconv.Atoi(f.Group); err == nil {
			h.Gid = gid
		}

		if err := tw.WriteHeader(h); err != nil {
			return nil, fmt.Errorf("writing header: %w", err)
		}

		if _, err := tw.Write([]byte(f.Content)); err != nil {
			return nil, fmt.Errorf("writing content: %w", err)
		}
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("closing writer: %w", err)
	}

	return buf, nil
}

// tarToFiles converts tar archive stream into list of container files.
func tarToFiles(rc io.Reader) ([]*types.File, error) {
	files := []*types.File{}
	buf := new(bytes.Buffer)
	tr := tar.NewReader(rc)

	for {
		header, err := tr.Next()
		if err == io.EOF { //nolint:errorlint
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed unpacking tar header: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		if _, err := buf.ReadFrom(tr); err != nil {
			return nil, fmt.Errorf("failed reading from tar archive: %w", err)
		}

		f := &types.File{
			User:    util.PickString(strconv.Itoa(header.Uid), header.Uname),
			Group:   util.PickString(strconv.Itoa(header.Gid), header.Gname),
			Content: buf.String(),
			Mode:    header.Mode,
		}

		files = append(files, f)
	}

	return files, nil
}

// Stat check if given paths exist on the container.
func (d *docker) Stat(id string, paths []string) (map[string]os.FileMode, error) {
	result := map[string]os.FileMode{}

	for _, p := range paths {
		s, err := d.cli.ContainerStatPath(d.ctx, id, p)
		if err != nil && !client.IsErrNotFound(err) {
			return nil, fmt.Errorf("statting path %q: %w", p, err)
		}

		if s.Name != "" {
			result[p] = s.Mode
		}
	}

	return result, nil
}

// Read reads files from container.
func (d *docker) Read(id string, srcPaths []string) ([]*types.File, error) {
	files := []*types.File{}

	for _, p := range srcPaths {
		rc, _, err := d.cli.CopyFromContainer(d.ctx, id, p)
		if err != nil && !client.IsErrNotFound(err) {
			return nil, fmt.Errorf("failed copying from container: %w", err)
		}

		// File does not exist.
		if rc == nil {
			continue
		}

		fs, err := tarToFiles(rc)
		if err != nil {
			return nil, fmt.Errorf("failed extracting file %s from archive: %w", p, err)
		}

		if err := rc.Close(); err != nil {
			return nil, fmt.Errorf("failed closing file: %w", err)
		}

		fs[0].Path = p

		files = append(files, fs[0])
	}

	return files, nil
}

// sanitizeImageName ensures, that given image name has tag in it's name.
// This is to ensure, that we can find the ID of the given image.
func sanitizeImageName(image string) string {
	if !strings.Contains(image, ":") {
		return fmt.Sprintf("%s:latest", image)
	}

	return image
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

	name := sanitizeImageName(image)

	for _, i := range images {
		for _, tag := range i.RepoTags {
			if tag == name {
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

	if _, err := io.Copy(ioutil.Discard, out); err != nil {
		return fmt.Errorf("failed to discard pulling messages: %w", err)
	}

	return out.Close()
}

// DefaultConfig returns Docker's runtime default configuration.
func DefaultConfig() *Config {
	return &Config{
		Host: client.DefaultDockerHost,
	}
}
