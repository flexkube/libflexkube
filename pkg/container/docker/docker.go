package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"

	"github.com/invidian/etcd-ariadnes-thread/pkg/container"
)

// Docker struct represents Docker container runtime
type Docker struct{}

// Start starts Docker container
func (d *Docker) Start(config *container.Config) error {
	fmt.Println("Starting docker container")
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return fmt.Errorf("Failed to create Docker client: %s", err)
	}
	list, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return errors.Wrap(err, "container list")
	}
	for _, c := range list {
		for _, n := range c.Names {
			fmt.Println("Checkng name:", n)
			if n == fmt.Sprintf("/%s", config.Name) {
				fmt.Println("Old container found! Removing...")
				if err := cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{}); err != nil {
					return errors.Wrap(err, "removing old container")
				}
			}
		}
	}

	_, err = cli.ImagePull(ctx, config.Image, types.ImagePullOptions{})
	if err != nil {
		return errors.Wrap(err, "pulling image")
	}

	dockerConfig := containertypes.Config{
		Image: config.Image,
		Cmd:   []string{"/bin/sh", "-c", "mount"},
	}
	hostConfig := containertypes.HostConfig{
		Mounts: []mount.Mount{},
	}

	fmt.Println("Creating container")
	c, err := cli.ContainerCreate(ctx, &dockerConfig, &hostConfig, &network.NetworkingConfig{}, config.Name)
	if err != nil {
		return errors.Wrap(err, "creating container")
	}
	ci, err := cli.ContainerInspect(ctx, c.ID)
	if err != nil {
		return errors.Wrap(err, "container inspect")
	}
	fmt.Println("Created container:", ci.Name)
	startOptions := types.ContainerStartOptions{}

	if err := cli.ContainerStart(ctx, c.ID, startOptions); err != nil {
		return errors.Wrap(err, "starting container")
	}
	return nil
}

// Stop stops Docker container
func (d *Docker) Stop(containerName string) error {
	return nil
}

// Status returns container status
func (d *Docker) Status(containerName string) (*container.Status, error) {
	return &container.Status{
		Image: "notimplemented",
	}, nil
}
