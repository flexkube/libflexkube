// Package docker implements runtime.Interface and runtime.Config interfaces
// by talking to Docker API.
package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
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
	// How long we wait when gracefully stopping the container before force-killing it.
	stopTimeoutSeconds = 30
)

// Config struct represents Docker container runtime configuration.
type Config struct {
	// Host is a Docker runtime URL. Usually 'unix:///run/docker.sock'. If empty
	// Docker's default URL will be used.
	Host string `json:"host,omitempty"`

	// ClientGetter allows to use custom Docker client.
	ClientGetter func(...client.Opt) (Client, error) `json:"-"`
}

// Client is a wrapper interface over
// https://godoc.org/github.com/docker/docker/client#ContainerAPIClient
// with the functions we use.
type Client interface {
	ContainerCreate(
		ctx context.Context,
		config *container.Config,
		hostConfig *container.HostConfig,
		networkingConfig *networktypes.NetworkingConfig,
		platform *v1.Platform,
		containerName string,
	) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, container string, options dockertypes.ContainerStartOptions) error
	ContainerStop(ctx context.Context, container string, options container.StopOptions) error
	ContainerInspect(ctx context.Context, container string) (dockertypes.ContainerJSON, error)
	ContainerRemove(ctx context.Context, container string, options dockertypes.ContainerRemoveOptions) error
	CopyFromContainer(
		ctx context.Context,
		container,
		srcPath string,
	) (io.ReadCloser, dockertypes.ContainerPathStat, error)
	CopyToContainer(
		ctx context.Context,
		container,
		path string,
		content io.Reader,
		options dockertypes.CopyToContainerOptions,
	) error
	ContainerStatPath(ctx context.Context, container, path string) (dockertypes.ContainerPathStat, error)
	ImageList(ctx context.Context, options dockertypes.ImageListOptions) ([]dockertypes.ImageSummary, error)
	ImagePull(ctx context.Context, ref string, options dockertypes.ImagePullOptions) (io.ReadCloser, error)
}

// docker struct is a struct, which can be used to manage Docker containers.
type docker struct {
	ctx context.Context //nolint:containedctx // Ignore until runtime interface supports context.
	cli Client
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

func (c *Config) getDockerClient() (Client, error) {
	opts := []client.Opt{
		client.WithVersion(defaults.DockerAPIVersion),
	}

	if c != nil && c.Host != "" {
		opts = append(opts, client.WithHost(c.Host))
	}

	if c.ClientGetter == nil {
		return client.NewClientWithOpts(opts...)
	}

	return c.ClientGetter(opts...)
}

// pullImageIfNotPresent pulls image if it's not already present on the host.
func (d *docker) pullImageIfNotPresent(image string) error {
	// Pull image to make sure it's available.
	// TODO make it configurable?
	id, err := d.imageID(image)
	if err != nil {
		return fmt.Errorf("checking for image presence: %w", err)
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

	for _, portMap := range ports {
		port, err := nat.NewPort(portMap.Protocol, strconv.Itoa(portMap.Port))
		if err != nil {
			return nil, nil, fmt.Errorf("mapping ports: %w", err)
		}

		if _, exists := portBindings[port]; !exists {
			portBindings[port] = []nat.PortBinding{}
		}

		portBindings[port] = append(portBindings[port], nat.PortBinding{
			HostIP:   portMap.IP,
			HostPort: strconv.Itoa(portMap.Port),
		})
		exposedPorts[port] = struct{}{}
	}

	return portBindings, exposedPorts, nil
}

// mounts converts container Mount to Docker mount type.
func mounts(containerMounts []types.Mount) []mount.Mount {
	dockerMounts := []mount.Mount{}

	for _, containerMount := range containerMounts {
		dockerMounts = append(dockerMounts, mount.Mount{
			Type:   "bind",
			Source: containerMount.Source,
			Target: containerMount.Target,
			// TODO validate!
			BindOptions: &mount.BindOptions{
				Propagation: mount.Propagation(containerMount.Propagation),
			},
		})
	}

	return dockerMounts
}

func convertContainerConfig(config *types.ContainerConfig) (*container.Config, *container.HostConfig, error) {
	// TODO That should be validated at ContainerConfig level!
	portBindings, exposedPorts, err := buildPorts(config.Ports)
	if err != nil {
		return nil, nil, fmt.Errorf("building ports: %w", err)
	}

	user := config.User
	if config.Group != "" {
		user = fmt.Sprintf("%s:%s", config.User, config.Group)
	}

	env := []string{}
	for k, v := range config.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Just structs required for starting container.
	dockerConfig := container.Config{
		Image:        config.Image,
		Cmd:          config.Args,
		Entrypoint:   config.Entrypoint,
		ExposedPorts: exposedPorts,
		User:         user,
		Env:          env,
	}
	hostConfig := container.HostConfig{
		Mounts:       mounts(config.Mounts),
		PortBindings: portBindings,
		Privileged:   config.Privileged,
		NetworkMode:  container.NetworkMode(config.NetworkMode),
		PidMode:      container.PidMode(config.PidMode),
		IpcMode:      container.IpcMode(config.IpcMode),
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	return &dockerConfig, &hostConfig, nil
}

// Start starts Docker container.
func (d *docker) Create(config *types.ContainerConfig) (string, error) {
	if err := d.pullImageIfNotPresent(config.Image); err != nil {
		return "", fmt.Errorf("pulling image: %w", err)
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
	timeout := stopTimeoutSeconds

	return d.cli.ContainerStop(d.ctx, id, container.StopOptions{
		Timeout: &timeout,
	})
}

// Status returns container status.
func (d *docker) Status(id string) (types.ContainerStatus, error) {
	containerStatus := types.ContainerStatus{
		ID: id,
	}

	status, err := d.cli.ContainerInspect(d.ctx, id)
	if err != nil {
		// If container is missing, return status with empty ID.
		if client.IsErrNotFound(err) {
			containerStatus.ID = ""

			return containerStatus, nil
		}

		return containerStatus, fmt.Errorf("inspecting container: %w", err)
	}

	containerStatus.Status = status.State.Status

	return containerStatus, nil
}

// Delete removes the container.
func (d *docker) Delete(id string) error {
	return d.cli.ContainerRemove(d.ctx, id, dockertypes.ContainerRemoveOptions{})
}

// Copy takes map of files and their content and copies it to the container using TAR archive.
//
// TODO Add support for base64 encoded content to support copying binary files.
func (d *docker) Copy(containerID string, files []*types.File) error {
	t, err := filesToTar(files)
	if err != nil {
		return fmt.Errorf("packing files to TAR archive: %w", err)
	}

	return d.cli.CopyToContainer(d.ctx, containerID, "/", t, dockertypes.CopyToContainerOptions{})
}

// filesToTar converts list of container files to tar archive format.
func filesToTar(files []*types.File) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tarWriter := tar.NewWriter(buf)

	for _, file := range files {
		header := &tar.Header{
			Name:    file.Path,
			Mode:    file.Mode,
			Size:    int64(len(file.Content)),
			ModTime: time.Now(),
			Uname:   file.User,
			Gname:   file.Group,
		}

		if uid, err := strconv.Atoi(file.User); err == nil {
			header.Uid = uid
		}

		if gid, err := strconv.Atoi(file.Group); err == nil {
			header.Gid = gid
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("writing header: %w", err)
		}

		if _, err := tarWriter.Write([]byte(file.Content)); err != nil {
			return nil, fmt.Errorf("writing content: %w", err)
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("closing writer: %w", err)
	}

	return buf, nil
}

// tarToFiles converts tar archive stream into list of container files.
func tarToFiles(rc io.Reader) ([]*types.File, error) {
	files := []*types.File{}
	buf := new(bytes.Buffer)
	tarReader := tar.NewReader(rc)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("unpacking tar header: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		if _, err := buf.ReadFrom(tarReader); err != nil {
			return nil, fmt.Errorf("reading from tar archive: %w", err)
		}

		file := &types.File{
			User:    util.PickString(strconv.Itoa(header.Uid), header.Uname),
			Group:   util.PickString(strconv.Itoa(header.Gid), header.Gname),
			Content: buf.String(),
			Mode:    header.Mode,
		}

		files = append(files, file)
	}

	return files, nil
}

// Stat check if given paths exist on the container.
func (d *docker) Stat(id string, paths []string) (map[string]os.FileMode, error) {
	result := map[string]os.FileMode{}

	for _, path := range paths {
		stat, err := d.cli.ContainerStatPath(d.ctx, id, path)
		if err != nil && !client.IsErrNotFound(err) {
			return nil, fmt.Errorf("statting path %q: %w", path, err)
		}

		if stat.Name != "" {
			result[path] = stat.Mode
		}
	}

	return result, nil
}

// Read reads files from container.
func (d *docker) Read(id string, srcPaths []string) ([]*types.File, error) {
	files := []*types.File{}

	for _, path := range srcPaths {
		stat, _, err := d.cli.CopyFromContainer(d.ctx, id, path)
		if err != nil && !client.IsErrNotFound(err) {
			return nil, fmt.Errorf("copying from container: %w", err)
		}

		// File does not exist.
		if stat == nil {
			continue
		}

		filesFromTar, err := tarToFiles(stat)
		if err != nil {
			return nil, fmt.Errorf("extracting file %s from archive: %w", path, err)
		}

		if err := stat.Close(); err != nil {
			return nil, fmt.Errorf("closing file: %w", err)
		}

		filesFromTar[0].Path = path

		files = append(files, filesFromTar[0])
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
		return "", fmt.Errorf("listing docker images: %w", err)
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
		return fmt.Errorf("pulling image: %w", err)
	}

	if _, err := io.Copy(io.Discard, out); err != nil {
		return fmt.Errorf("discarding pulling messages: %w", err)
	}

	return out.Close()
}

// DefaultConfig returns Docker's runtime default configuration.
func DefaultConfig() *Config {
	return &Config{
		Host: client.DefaultDockerHost,
	}
}
