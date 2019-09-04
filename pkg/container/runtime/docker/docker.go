package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container/runtime"
	"github.com/invidian/etcd-ariadnes-thread/pkg/defaults"
)

const runtimeName = "docker"

func init() {
	runtime.Register(runtimeName)
}

// Docker struct represents Docker container runtime
type Docker struct{}

type docker struct {
	ctx context.Context
	cli *client.Client
}

// New validates Docker runtime configuration and returns configured
// runtime client.
func New(d *Docker) (*docker, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion(defaults.DockerAPIVersion))
	if err != nil {
		return nil, errors.Wrap(err, "creating Docker client")
	}
	return &docker{
		ctx: context.Background(),
		cli: cli,
	}, nil
}

// Start starts Docker container
//
// This should be generic, so it can be used to start any kind of containers!
//
// TODO figure out how to do that on remote machine with SSH
func (d *docker) Create(config *runtime.Config) (string, error) {
	// Pull image to make sure it's available.
	// TODO make it configurable?
	if _, err := d.cli.ImagePull(d.ctx, config.Image, types.ImagePullOptions{}); err != nil {
		return "", errors.Wrap(err, "pulling image")
	}

	// Just structs required for starting container.
	dockerConfig := containertypes.Config{
		Image: config.Image,
	}
	hostConfig := containertypes.HostConfig{
		Mounts: []mount.Mount{},
	}

	// Create container
	c, err := d.cli.ContainerCreate(d.ctx, &dockerConfig, &hostConfig, &network.NetworkingConfig{}, config.Name)
	if err != nil {
		return "", errors.Wrap(err, "creating container")
	}

	return c.ID, nil
}

// Start starts Docker container
//
// This should be generic, so it can be used to start any kind of containers!
//
// TODO figure out how to do that on remote machine with SSH
func (d *docker) Start(ID string) error {
	return d.cli.ContainerStart(d.ctx, ID, types.ContainerStartOptions{})
}

// Stop stops Docker container
func (d *docker) Stop(ID string) error {
	timeout := time.Duration(30) * time.Second
	return d.cli.ContainerStop(d.ctx, ID, &timeout)
}

// Status returns container status
func (d *docker) Status(ID string) (*runtime.Status, error) {
	status, err := d.cli.ContainerInspect(d.ctx, ID)
	if err != nil {
		// If container is missing, return no status
		if client.IsErrNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "inspecting container failed")
	}

	return &runtime.Status{
		Image:  status.Image,
		ID:     ID,
		Name:   status.Name,
		Status: status.State.Status,
	}, nil
}

// Delete removes the container
func (d *docker) Delete(ID string) error {
	return d.cli.ContainerRemove(d.ctx, ID, types.ContainerRemoveOptions{})
}
