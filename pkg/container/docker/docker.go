package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container"
)

// Docker struct represents Docker container runtime
type Docker struct {
	ctx context.Context
	cli *client.Client
}

func New() (*Docker, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "creating Docker client")
	}
	return &Docker{
		ctx: context.Background(),
		cli: cli,
	}, nil
}

func (d *Docker) validate() error {
	if d.cli == nil || d.ctx == nil {
		return fmt.Errorf("container runtime is not properly intialized")
	}

	return nil
}

// Start starts Docker container
//
// This should be generic, so it can be used to start any kind of containers!
//
// TODO figure out how to do that on remote machine with SSH
func (d *Docker) Create(config *container.Config) (string, error) {
	if err := d.validate(); err != nil {
		return "", errors.Wrap(err, "creating container failed")
	}
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
func (d *Docker) Start(ID string) error {
	if err := d.validate(); err != nil {
		return errors.Wrap(err, "creating container failed")
	}

	// And start requested container
	return d.cli.ContainerStart(d.ctx, ID, types.ContainerStartOptions{})
}

// Stop stops Docker container
func (d *Docker) Stop(ID string) error {
	if err := d.validate(); err != nil {
		return errors.Wrap(err, "creating container failed")
	}

	timeout := time.Duration(30) * time.Second
	return d.cli.ContainerStop(d.ctx, ID, &timeout)
}

// Status returns container status
func (d *Docker) Status(ID string) (*container.Status, error) {
	if err := d.validate(); err != nil {
		return nil, errors.Wrap(err, "creating container failed")
	}

	status, err := d.cli.ContainerInspect(d.ctx, ID)
	if err != nil {
		// If container is missing, return no status
		if client.IsErrNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "inspecting container failed")
	}

	return &container.Status{
		Image:  status.Image,
		ID:     ID,
		Name:   status.Name,
		Status: status.State.Status,
	}, nil
}

// Delete removes the container
func (d *Docker) Delete(ID string) error {
	if err := d.validate(); err != nil {
		return errors.Wrap(err, "creating container failed")
	}

	return d.cli.ContainerRemove(d.ctx, ID, types.ContainerRemoveOptions{})
}
